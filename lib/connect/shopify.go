package connect

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	db "upper.io/db.v3"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
	"github.com/goware/geotools"
	"github.com/goware/jwtauth"
	"github.com/goware/lg"
	"github.com/pkg/errors"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/lib/token"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"

	"golang.org/x/oauth2"
)

type Shopify struct {
	//*oauth2.Config
	clientID     string
	clientSecret string
	redirectURL  string
	webhookURL   string
}

var (
	SH *Shopify
)

func SetupShopify(conf Config) {
	SH = &Shopify{
		clientID:     conf.AppId,
		clientSecret: conf.AppSecret,
		redirectURL:  conf.OAuthCallback,
		webhookURL:   conf.WebhookURL,
	}
}

func (s *Shopify) AuthCodeURL(ctx context.Context) string {
	place := ctx.Value("place").(*data.Place)
	config := s.getConfig(place.ShopifyID)
	t, _ := token.Encode(jwtauth.Claims{"place_shop_id": place.ShopifyID})
	return config.AuthCodeURL(t.Raw, oauth2.AccessTypeOffline)
}

// callback from initiating AuthCodeURL
func (s *Shopify) OAuthCb(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	// TODO: check HMAC signature
	//code := q.Get("code")

	token, err := token.Decode(q.Get("state"))
	if err != nil {
		ws.Respond(w, http.StatusBadRequest, err)
		return
	}

	shopID, ok := token.Claims["place_shop_id"].(string)
	if !ok {
		ws.Respond(w, http.StatusBadRequest, ErrInvalidState)
		return
	}

	if fmt.Sprintf("%s.myshopify.com", shopID) != q.Get("shop") {
		ws.Respond(w, http.StatusBadRequest, ErrMismathShop)
		return
	}

	tok, err := s.Exchange(shopID, r)
	if err != nil {
		ws.Respond(w, http.StatusBadRequest, err)
		return
	}

	creds := &data.ShopifyCred{
		AccessToken: tok.AccessToken,
		ApiURL:      fmt.Sprintf("https://%s.myshopify.com", shopID),
	}

	if err := s.finalizeCallback(r.Context(), shopID, creds); err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	ws.Respond(w, http.StatusCreated, "shopify connected.")
}

func (s *Shopify) finalizeCallback(ctx context.Context, shopID string, creds *data.ShopifyCred) error {
	api := shopify.NewClient(nil, creds.AccessToken)
	api.BaseURL, _ = url.Parse(creds.ApiURL)

	// fetch place data from shopify
	shop, _, err := api.Shop.Get(ctx)
	if err != nil {
		return err
	}

	place, err := data.DB.Place.FindByShopifyID(shopID)
	if err != nil && err != db.ErrNoMoreRows {
		return errors.Wrap(err, "failed to look up place")
	}

	if place == nil {
		// find locale from latlng
		latlng := s2.LatLngFromDegrees(shop.Longitude, shop.Longitude)
		origin := s2.CellIDFromLatLng(latlng).Parent(15) // 16 for more detail?
		// Find the reach of cells
		cond := db.Cond{
			"cell_id >=": int(origin.RangeMin()),
			"cell_id <=": int(origin.RangeMax()),
		}
		cells, err := data.DB.Cell.FindAll(cond)
		if err != nil {
			return err
		}

		// Find the minimum distance cell
		min := s1.InfAngle()
		var localeID int64
		for _, c := range cells {
			cell := s2.CellID(c.CellID)
			d := latlng.Distance(cell.LatLng())
			if d < min {
				min = d
				localeID = c.LocaleID
			}
		}
		u, _ := url.Parse(shop.Domain)
		u.Scheme = "https"
		place = &data.Place{
			ShopifyID: shopID,
			LocaleID:  localeID,
			Geo:       *geotools.NewPointFromLatLng(shop.Latitude, shop.Longitude),
			Name:      shop.Name,
			Address:   fmt.Sprintf("%s, %s", shop.Address1, shop.City),
			Phone:     shop.Phone,
			Website:   u.String(),
		}
	}
	if err := data.DB.Place.Save(place); err != nil {
		return errors.Wrap(err, "failed to save place")
	}

	// save authorization
	creds.PlaceID = place.ID
	if err := data.DB.ShopifyCred.Save(creds); err != nil {
		return errors.Wrap(err, "failed to save cred")
	}

	// create the webhook
	wh := &shopify.WebhookRequest{
		&shopify.Webhook{
			Topic:   shopify.TopicProductsCreate,
			Address: s.webhookURL,
			Format:  "json",
		},
	}
	_, _, err = api.Webhook.Create(ctx, wh)
	if err != nil {
		lg.Alert(errors.Wrap(err, "shopify webhook"))
	}

	// fetch the product list
	productList, _, err := api.Product.List(ctx)
	if err != nil {
		return errors.Wrap(err, "shopify list products")
	}

	// initial sync up
	for _, p := range productList {
		imgUrl, _ := url.Parse(p.Image.Src)
		imgUrl.Scheme = "https"

		product := &data.Product{
			PlaceID:     place.ID,
			ExternalID:  p.Handle,
			Title:       p.Title,
			Description: p.BodyHTML,
			ImageUrl:    imgUrl.String(),
		}
		product.ParseTags(p.Tags, p.ProductType, p.Vendor)
		var promos []*data.Promo
		for _, v := range p.Variants {
			now := time.Now().UTC()
			start := now.Add(1 * time.Minute)
			end := now.Add(30 * 24 * time.Hour)
			price, _ := strconv.ParseFloat(v.Price, 64)
			promo := &data.Promo{
				PlaceID:     place.ID,
				ProductID:   product.ID,
				Type:        data.PromoTypePrice,
				OfferID:     v.ID,
				Status:      data.PromoStatusActive,
				Description: v.Title,
				UserID:      0, // admin
				Limits:      int64(v.InventoryQuantity),
				Etc: data.PromoEtc{
					Price: price,
					Sku:   v.Sku,
				},
				StartAt: &start,
				EndAt:   &end, // 1 month
			}
			promos = append(promos, promo)
		}

		if err := data.DB.Product.Save(product); err != nil {
			return errors.Wrap(err, "failed to save promotion")
		}

		for _, v := range promos {
			v.ProductID = product.ID
			if err := data.DB.Promo.Save(v); err != nil {
				lg.Warn(errors.Wrap(err, "failed to save promotion"))
				continue
			}
		}
	}

	return nil
}

func (s *Shopify) Exchange(shopID string, r *http.Request) (*oauth2.Token, error) {
	code := r.URL.Query().Get("code")

	config := s.getConfig(shopID)
	return config.Exchange(r.Context(), code)
}

// NOTE: added ".myshopify.com" to oauth2 vendored lib
func (s *Shopify) getConfig(shopifyID string) *oauth2.Config {
	shopUrl := fmt.Sprintf("https://%s.myshopify.com", shopifyID)
	return &oauth2.Config{
		ClientID:     s.clientID,
		ClientSecret: s.clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/admin/oauth/authorize", shopUrl),
			TokenURL: fmt.Sprintf("%s/admin/oauth/access_token", shopUrl),
		},
		RedirectURL: s.redirectURL,
		Scopes:      []string{"read_products"},
	}
}

// NOTE: this doesn't work unless we implement our own
// http transport and token and make it play nice with oauth2 lib
func (s *Shopify) ClientFromCred(r *http.Request) *http.Client {
	ctx := r.Context()
	cred := ctx.Value("creds").(*data.ShopifyCred)

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cred.AccessToken},
	)

	return oauth2.NewClient(ctx, ts)
}
