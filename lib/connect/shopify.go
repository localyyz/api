package connect

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/goware/jwtauth"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/token"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"

	"golang.org/x/oauth2"
)

type Shopify struct {
	//*oauth2.Config
	clientID     string
	clientSecret string
	redirectURL  string
}

var (
	SH *Shopify
)

func SetupShopify(conf Config) {
	SH = &Shopify{
		clientID:     conf.AppId,
		clientSecret: conf.AppSecret,
		redirectURL:  conf.OAuthCallback,
	}
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

// callback from initiating AuthCodeURL
func (s *Shopify) OAuthCb(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	// TODO: check HMAC signature

	//code := q.Get("code")
	state := q.Get("state")
	shop := q.Get("shop")

	token, err := token.Decode(state)
	if err != nil {
		ws.Respond(w, http.StatusBadRequest, err)
		return
	}

	pId, ok := token.Claims["place_id"].(string)
	if !ok {
		ws.Respond(w, http.StatusBadRequest, ErrInvalidState)
		return
	}

	placeId, _ := strconv.ParseInt(pId, 10, 64)
	place, err := data.DB.Place.FindByID(placeId)

	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	if fmt.Sprintf("%s.myshopify.com", place.ShopifyID) != shop {
		ws.Respond(w, http.StatusBadRequest, "")
		return
	}

	tok, err := s.Exchange(place, r)
	if err != nil {
		ws.Respond(w, http.StatusBadRequest, err)
		return
	}

	// save authorization
	cred := &data.ShopifyCred{
		PlaceID:     place.ID,
		AccessToken: tok.AccessToken,
		ApiURL:      fmt.Sprintf("https://%s.myshopify.com", place.ShopifyID),
	}
	if err := data.DB.ShopifyCred.Save(cred); err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	ws.Respond(w, http.StatusCreated, "shopify connected.")
}

func (s *Shopify) createWebhook(cred *data.ShopifyCred) error {
	return nil
}

func (s *Shopify) AuthCodeURL(r *http.Request) string {
	place := r.Context().Value("place").(*data.Place)

	config := s.getConfig(place.ShopifyID)
	t, _ := token.Encode(jwtauth.Claims{"place_id": fmt.Sprintf("%d", place.ID)})
	return config.AuthCodeURL(t.Raw, oauth2.AccessTypeOffline)
}

func (s *Shopify) Exchange(place *data.Place, r *http.Request) (*oauth2.Token, error) {
	code := r.URL.Query().Get("code")

	config := s.getConfig(place.ShopifyID)
	return config.Exchange(r.Context(), code)
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
