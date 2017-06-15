package connect

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	db "upper.io/db.v3"

	"github.com/goware/geotools"
	"github.com/goware/jwtauth"
	"github.com/goware/lg"
	"github.com/pkg/errors"
	"github.com/pressly/chi/render"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/lib/token"

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
		render.Status(r, http.StatusBadRequest)
		render.Respond(w, r, err)
		return
	}

	shopID, ok := token.Claims["place_shop_id"].(string)
	if !ok {
		render.Status(r, http.StatusBadRequest)
		render.Respond(w, r, ErrInvalidState)
		return
	}

	if fmt.Sprintf("%s.myshopify.com", shopID) != q.Get("shop") {
		render.Status(r, http.StatusBadRequest)
		render.Respond(w, r, ErrMismathShop)
		return
	}

	tok, err := s.Exchange(shopID, r)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.Respond(w, r, err)
		return
	}

	creds := &data.ShopifyCred{
		AccessToken: tok.AccessToken,
		ApiURL:      fmt.Sprintf("https://%s.myshopify.com", shopID),
	}

	if err := s.finalizeCallback(r.Context(), shopID, creds); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.Respond(w, r, err)
		return
	}

	render.Status(r, http.StatusCreated)
	render.Respond(w, r, "shopify connected.")
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
		locale, err := data.DB.Locale.FromLatLng(shop.Latitude, shop.Longitude)
		if err != nil {
			return err
		}

		u, _ := url.Parse(shop.Domain)
		u.Scheme = "https"
		place = &data.Place{
			ShopifyID: shopID,
			LocaleID:  locale.ID,
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
		var promos []*data.Promo
		for _, v := range p.Variants {
			price, _ := strconv.ParseFloat(v.Price, 64)
			promo := &data.Promo{
				PlaceID:     place.ID,
				ProductID:   product.ID,
				OfferID:     v.ID,
				Status:      data.PromoStatusActive,
				Description: v.Title,
				UserID:      0, // admin
				Limits:      int64(v.InventoryQuantity),
				Etc: data.PromoEtc{
					Price: price,
					Sku:   v.Sku,
				},
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

		tags := product.ParseTags(p.Tags, p.ProductType, p.Vendor)
		q := data.DB.InsertInto("product_tags").Columns("product_id", "value")
		b := q.Batch(len(tags))
		go func() {
			defer b.Done()
			for _, t := range tags {
				b.Values(product.ID, t)
			}
		}()
		if err := b.Wait(); err != nil {
			lg.Warn(err)
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
