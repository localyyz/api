package connect

import (
	"fmt"
	"net/http"

	"github.com/goware/jwtauth"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/token"

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
