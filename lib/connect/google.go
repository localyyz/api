package connect

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"regexp"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type Google struct {
	appId     string
	appSecret string

	certs      certs
	certExpiry time.Time
}

type certs map[string]*rsa.PublicKey

type certKey struct {
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	Kid string `json:"Kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

var (
	// Issuers is the allowed oauth token issuers
	Issuers = []string{
		"accounts.google.com",
		"https://accounts.google.com",
	}

	// The URL that provides public certificates for verifying ID tokens issued
	// by Google's OAuth 2.0 authorization server.
	GoogleOauth2CertsURL = "https://www.googleapis.com/oauth2/v3/certs"

	// instance
	GN *Google
)

func SetupGoogle(conf Config) *Google {
	GN = &Google{
		appId:     conf.AppId,
		appSecret: conf.AppSecret,
	}
	return GN
}

// parse response header for cache control
func parseCacheControl(header http.Header) (int64, error) {
	cacheAge := int64(7200) // Set default cacheAge to 2 hours

	// cache-control: public, max-age=23814, must-revalidate, no-transform
	cacheControl := header.Get("cache-control")
	if len(cacheControl) == 0 {
		return cacheAge, nil
	}

	re := regexp.MustCompile("max-age=([0-9]*)")
	match := re.FindAllStringSubmatch(cacheControl, -1)
	if len(match) > 0 {
		if len(match[0]) == 2 {
			maxAge := match[0][1]
			maxAgeInt, err := strconv.ParseInt(maxAge, 10, 64)
			if err != nil {
				return cacheAge, err
			}
			cacheAge = maxAgeInt
		}
	}

	return cacheAge, nil
}

func (g *Google) fetchCerts() (certs, error) {
	if g.certs != nil && time.Now().Before(g.certExpiry) {
		return g.certs, nil
	}
	resp, err := http.Get(GoogleOauth2CertsURL)
	if err != nil {
		return nil, err
	}

	cacheAge, err := parseCacheControl(resp.Header)
	if err != nil {
		return nil, err
	}
	g.certExpiry = time.Now().Add(time.Second * time.Duration(cacheAge))

	var keyWrapper struct {
		Keys []certKey `json:"keys"`
	}
	err = json.NewDecoder(resp.Body).Decode(&keyWrapper)
	if err != nil {
		return nil, err
	}

	g.certs = map[string]*rsa.PublicKey{}
	for _, key := range keyWrapper.Keys {
		if key.Use == "sig" && key.Kty == "RSA" {
			n, err := base64.RawURLEncoding.DecodeString(key.N)
			if err != nil {
				return nil, err
			}
			e, err := base64.RawURLEncoding.DecodeString(key.E)
			if err != nil {
				return nil, err
			}
			ei := big.NewInt(0).SetBytes(e).Int64()
			if err != nil {
				return nil, err
			}
			g.certs[key.Kid] = &rsa.PublicKey{
				N: big.NewInt(0).SetBytes(n),
				E: int(ei),
			}
		}
	}

	return g.certs, nil
}

func (g *Google) tokenKeyFn(token *jwt.Token) (interface{}, error) {
	certs, err := g.fetchCerts()
	if err != nil {
		return false, err
	}
	key, ok := certs[token.Header["kid"].(string)]
	if !ok {
		return nil, ErrTokenPubKey
	}
	return key, nil
}

func (g *Google) VerifyToken(token string) (bool, error) {
	parser := &jwt.Parser{}
	jwtToken, err := parser.Parse(token, g.tokenKeyFn)
	if err != nil {
		return false, err
	}
	if err := jwtToken.Claims.Valid(); err != nil {
		return false, err
	}
	claims, _ := jwtToken.Claims.(*jwt.StandardClaims)
	if verify := claims.VerifyAudience(g.appId, true); !verify {
		return false, ErrTokenAudience
	}
	return false, nil
}
