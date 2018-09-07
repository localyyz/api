package connect

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	db "upper.io/db.v3"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
	"github.com/pkg/errors"
	"github.com/pressly/lg"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/lib/token"

	"golang.org/x/oauth2"
)

type Shopify struct {
	//*oauth2.Config
	appName      string
	clientID     string
	clientSecret string
	redirectURL  string
	webhookURL   string
}

var (
	SH     *Shopify
	Scopes = []string{
		"read_content",
		"read_products",
		"read_product_listings",
		"read_collection_listings",
		"read_checkouts",
		"write_checkouts",
		"read_price_rules",
		"write_price_rules",
	}
	WebhookTopics = []shopify.Topic{
		shopify.TopicShopUpdate,
		shopify.TopicAppUninstalled,
		shopify.TopicProductListingsAdd,
		shopify.TopicProductListingsUpdate,
		shopify.TopicProductListingsRemove,
		shopify.TopicCollectionListingsAdd,
		shopify.TopicCollectionListingsUpdate,
		shopify.TopicCollectionListingsRemove,
		shopify.TopicCheckoutsUpdate,
	}
)

func SetupShopify(conf Config) *Shopify {
	SH = &Shopify{
		appName:      conf.AppName,
		clientID:     conf.AppId,
		clientSecret: conf.AppSecret,
		redirectURL:  conf.OAuthCallback,
		webhookURL:   conf.WebhookURL,
	}
	return SH
}

// read only access to pp name
func (s *Shopify) AppName() string {
	return s.appName
}

// read only access to client id
func (s *Shopify) ClientID() string {
	return s.clientID
}

func (s *Shopify) RegisterWebhooks(ctx context.Context, place *data.Place) {
	creds, err := data.DB.ShopifyCred.FindByPlaceID(place.ID)
	if err != nil {
		lg.Warnf("register webhook: %s(id:%v) %+v", place.Name, place.ID, err)
		return
	}
	sh := shopify.NewClient(nil, creds.AccessToken)
	sh.BaseURL, _ = url.Parse(creds.ApiURL)
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
			lg.Warnf("register webhook: %s(id:%v) %s with %+v", place.Name, place.ID, topic, err)
			continue
		}
		if err := data.DB.Webhook.Save(&data.Webhook{
			PlaceID:    place.ID,
			Topic:      string(topic),
			ExternalID: int64(wh.ID),
		}); err != nil {
			lg.Warnf("register webhook: %s (id:%v) %s with %+v", place.Name, place.ID, topic, err)
		}
	}
	lg.Warnf("registered webhooks for place(%d)", place.ID)
}

func (s *Shopify) RegisterReturnPolicy(ctx context.Context, place *data.Place) {
	client, err := GetShopifyClient(place.ID)
	if err != nil {
		lg.Warnf("failed to fetch client for place(%d): %v", place.ID, err)
		return
	}

	policies, _, err := client.Policy.List(ctx)
	if err != nil {
		lg.Alertf("connect %s (id: %v): failed to fetch policy %+v", place.Name, place.ID, err)
		return
	}
	for _, p := range policies {
		if p.Title == shopify.PolicyRefund {
			place.ReturnPolicy.Description = p.Body
			place.ReturnPolicy.URL = p.URL
		}
	}
	err = data.DB.Place.Save(place)
	if err != nil {
		lg.Warnf("failed to save return policy place(%d): %v", place.ID, err)
		return
	}
	lg.Warnf("registered return policy for place(%d)", place.ID)
}

func (s *Shopify) RegisterShippingPolicy(ctx context.Context, place *data.Place) {
	client, err := GetShopifyClient(place.ID)
	if err != nil {
		lg.Warnf("failed to fetch client for place(%d): %v", place.ID, err)
		return
	}

	// fetch shipping policies
	zones, _, _ := client.ShippingZone.List(ctx)
	for _, z := range zones {
		for _, c := range z.Countries {
			if c.Code != "US" && c.Code != "CA" {
				continue
			}
			sz := data.ShippingZone{
				PlaceID:    place.ID,
				Name:       z.Name,
				ExternalID: z.ID,
				Country:    strings.ToLower(c.Name),
			}
			for _, r := range c.Provinces {
				sz.Regions = append(sz.Regions, data.Region{
					Region:     r.Name,
					RegionCode: r.Code,
				})
			}

			for _, wz := range z.WeightBasedShippingRates {
				wpz := sz
				wpz.Type = data.ShippingZoneTypeByWeight
				wpz.Description = wz.Name
				wpz.WeightLow = wz.WeightLow
				wpz.WeightHigh = wz.WeightHigh
				wpz.Price, _ = strconv.ParseFloat(wz.Price, 64)
				data.DB.ShippingZone.Save(&wpz)
			}

			for _, pz := range z.PriceBasedShippingRates {
				ppz := sz
				ppz.Type = data.ShippingZoneTypeByPrice
				ppz.Description = pz.Name
				ppz.SubtotalLow, _ = strconv.ParseFloat(pz.MinOrderSubtotal, 64)
				ppz.SubtotalHigh, _ = strconv.ParseFloat(pz.MaxOrderSubtotal, 64)
				ppz.Price, _ = strconv.ParseFloat(pz.Price, 64)
				data.DB.ShippingZone.Save(&ppz)
			}
		}

	}

	lg.Warnf("registered return policy for place(%d)", place.ID)
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

	decodeToken, err := token.Decode(q.Get("state"))
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.Respond(w, r, err)
		return
	}

	// TODO: do this with context
	claims, ok := decodeToken.Claims.(jwt.MapClaims)
	if !ok {
		render.Status(r, http.StatusBadRequest)
		render.Respond(w, r, err)
		return
	}

	shopID, ok := claims["place_shop_id"].(string)
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

	oauthToken, err := s.Exchange(shopID, r)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.Respond(w, r, err)
		return
	}

	creds := &data.ShopifyCred{
		AccessToken: oauthToken.AccessToken,
		ApiURL:      fmt.Sprintf("https://%s.myshopify.com", shopID),
		Status:      data.ShopifyCredStatusActive,
	}

	if err := s.finalizeCallback(r.Context(), shopID, creds); err != nil {
		lg.Alertf("connect finalize %s: failed with %+v", shopID, err)
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
		u, _ := url.Parse(shop.Domain)
		u.Scheme = "https"
		place = &data.Place{
			ShopifyID:   shopID,
			Plan:        shop.PlanName,
			Name:        shop.Name,
			Address:     fmt.Sprintf("%s, %s", shop.Address1, shop.City),
			Phone:       shop.Phone,
			Website:     u.String(),
			Currency:    shop.Currency,
			PlanEnabled: true,
		}
	}

	// check place status, if already active, skip the rest
	if place.Status != data.PlaceStatusActive {
		// upgrade place status to "waiting for agreement"
		place.Status = data.PlaceStatusWaitAgreement
		place.Gender = data.PlaceGender(data.ProductGenderUnisex)

		// create a place holder checkout for the account id
		place.PaymentMethods = []*data.PaymentMethod{}
		if checkout, _, _ := sh.Checkout.Create(ctx, nil); checkout != nil && len(checkout.ShopifyPaymentAccountID) != 0 {
			// NOTE: for now the id returned on checkout is stripe specific
			place.PaymentMethods = append(place.PaymentMethods, &data.PaymentMethod{Type: "stripe", ID: checkout.ShopifyPaymentAccountID})
		}
		var merchantApproval *data.MerchantApproval
		// save the place!
		if err := data.DB.Place.Save(place); err != nil {
			return errors.Wrap(err, "failed to save place")
		}
		// if merchant approval is not nil, save
		if merchantApproval != nil {
			merchantApproval.PlaceID = place.ID
			data.DB.MerchantApproval.Save(merchantApproval)
		}
		// create the webhook
		wh, _, err := sh.Webhook.Create(
			ctx,
			&shopify.WebhookRequest{
				&shopify.Webhook{
					Topic:   shopify.TopicAppUninstalled,
					Address: s.webhookURL,
					Format:  "json",
				},
			})
		if err != nil {
			lg.Warnf("connect %s (id: %v): webhook AppUninstall with %+v", place.Name, place.ID, err)
			return nil
		}
		err = data.DB.Webhook.Save(&data.Webhook{
			PlaceID:    place.ID,
			Topic:      string(shopify.TopicAppUninstalled),
			ExternalID: int64(wh.ID),
		})
		if err != nil {
			lg.Warnf("connect %s (id: %v): webhook AppUninstall with %+v", place.Name, place.ID, err)
		}
	}

	// save authorization
	creds.PlaceID = place.ID
	// check if creds already exists, and fill the id and make update
	if dbCreds, _ := data.DB.ShopifyCred.FindByPlaceID(place.ID); dbCreds != nil {
		creds.ID = dbCreds.ID
	}
	if err := data.DB.ShopifyCred.Save(creds); err != nil {
		return errors.Wrap(err, "failed to save cred")
	}

	SL.Notify("store", fmt.Sprintf("%s (id: %d - %s) just connected!", place.Name, place.ID, place.Plan))
	return nil
}

// this is called internally by the approval tool
func (s *Shopify) Finalize(w http.ResponseWriter, r *http.Request) {

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

func GetShopifyClient(merchantID int64) (*shopify.Client, error) {
	// getting the shopify cred
	cred, err := data.DB.ShopifyCred.FindOne(db.Cond{"place_id": merchantID})
	if err != nil {
		return nil, err
	}

	// creating the client
	client := shopify.NewClient(nil, cred.AccessToken)
	client.BaseURL, err = url.Parse(cred.ApiURL)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func init() {
	// NOTE: register broken oauth2
	oauth2.RegisterBrokenAuthHeaderProvider(".myshopify.com")
}
