package connect

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	db "upper.io/db.v3"

	"github.com/go-chi/render"
	"github.com/goware/geotools"
	"github.com/goware/jwtauth"
	"github.com/pkg/errors"
	"github.com/pressly/lg"

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
	SH     *Shopify
	Scopes = []string{
		"read_products",
		"read_product_listings",
		"read_collection_listings",
		"read_checkouts",
		"write_checkouts",
		"read_price_rules",
	}
	WebhookTopics = []shopify.Topic{
		shopify.TopicAppUninstalled,
		shopify.TopicProductListingsAdd,
		shopify.TopicProductListingsUpdate,
		shopify.TopicProductListingsRemove,
		shopify.TopicCollectionListingsAdd,
		shopify.TopicCollectionListingsUpdate,
		shopify.TopicCollectionListingsRemove,
	}

	TrialEndDate time.Time
)

func SetupShopify(conf Config) *Shopify {
	SH = &Shopify{
		clientID:     conf.AppId,
		clientSecret: conf.AppSecret,
		redirectURL:  conf.OAuthCallback,
		webhookURL:   conf.WebhookURL,
	}
	TrialEndDate, _ = time.Parse("2006-01-02", "2018-05-01")
	return SH
}

// read only access to client id
func (s *Shopify) ClientID() string {
	return s.clientID
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
		Status:      data.ShopifyCredStatusActive,
	}

	if err := s.finalizeCallback(r.Context(), shopID, creds); err != nil {
		lg.Alertf("connect %s: failed with %+v", shopID, err)
		render.Status(r, http.StatusInternalServerError)
		render.Respond(w, r, err)
		return
	}

	// redirect the user to shopify admin
	adminUrl := fmt.Sprintf("https://%s.myshopify.com/admin/apps/localyyz", shopID)
	http.Redirect(w, r, adminUrl, http.StatusTemporaryRedirect)
}

func (s *Shopify) finalizeCallback(ctx context.Context, shopID string, creds *data.ShopifyCred) error {
	sh := shopify.NewClient(nil, creds.AccessToken)
	sh.BaseURL, _ = url.Parse(creds.ApiURL)

	// fetch place data from shopify
	shop, _, err := sh.Shop.Get(ctx)
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
	// upgrade place status to "waiting for agreement"
	place.Status = data.PlaceStatusWaitAgreement

	// create the recurring billing, set at 120.00
	shopifyBilling := &shopify.Billing{
		Type:      shopify.BillingTypeRecurring,
		Name:      "Localyyz Subscription",
		Price:     "120.00",
		ReturnUrl: fmt.Sprintf("https://%s.myshopify.com/admin/apps/localyyz", shopID),
		TrialDays: int64(time.Until(TrialEndDate).Hours() / 24),
	}
	_, _, err = sh.Billing.Create(ctx, shopifyBilling)
	if err != nil {
		return errors.Wrapf(err, "failed to create billing for %s", shopID)
	}
	place.Billing = data.PlaceBilling{Billing: shopifyBilling}

	// save the place!
	if err := data.DB.Place.Save(place); err != nil {
		return errors.Wrap(err, "failed to save place")
	}

	// save authorization
	creds.PlaceID = place.ID
	// check if creds already exists, and fill the id
	if dbCreds, _ := data.DB.ShopifyCred.FindByPlaceID(place.ID); dbCreds != nil {
		creds.ID = dbCreds.ID
	}
	if err := data.DB.ShopifyCred.Save(creds); err != nil {
		return errors.Wrap(err, "failed to save cred")
	}

	// create the webhook
	for _, topic := range WebhookTopics {
		wh, _, err := sh.Webhook.Create(
			ctx,
			&shopify.WebhookRequest{
				&shopify.Webhook{
					Topic:   topic,
					Address: s.webhookURL,
					Format:  "json",
				},
			})
		if err != nil {
			lg.Alertf("connect %s (id: %v): failed to create shopify %s webhook with %+v", topic, place.ID, place.Name, err)
			continue
		}
		err = data.DB.Webhook.Save(&data.Webhook{
			PlaceID:    place.ID,
			Topic:      string(topic),
			ExternalID: int64(wh.ID),
		})
		if err != nil {
			lg.Alertf("connect %s (id: %v): failed to save shopify %s webhook with %+v", topic, place.ID, place.Name, err)
			continue
		}
	}

	SL.Notify("store", fmt.Sprintf("%s (id: %v) just connected!", place.Name, place.ID))
	return nil
}

func (s *Shopify) Exchange(shopID string, r *http.Request) (*oauth2.Token, error) {
	code := r.URL.Query().Get("code")

	config := s.getConfig(shopID)
	return config.Exchange(r.Context(), code)
}

// NOTE: added ".myshopify.com" to oauth2 vendored lib (internal/token.go -> brokenAuthHeaderDomains)
// NOTE: changed AuthCodeURL scope to be comma deliminated
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
		Scopes:      []string{strings.Join(Scopes, ",")},
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

func (s *Shopify) VerifySignature(sig []byte, query string) bool {
	mac := hmac.New(sha256.New, []byte(s.clientSecret))
	// query unescape
	uu, _ := url.QueryUnescape(query)
	mac.Write([]byte(uu))

	src := mac.Sum(nil)
	// hex encode
	expectedSig := make([]byte, hex.EncodedLen(len(src)))
	hex.Encode(expectedSig, src)

	return hmac.Equal(sig, expectedSig)
}

func init() {
	// NOTE: register broken oauth2
	oauth2.RegisterBrokenAuthHeaderProvider(".myshopify.com")
}
