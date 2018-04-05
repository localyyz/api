package token

import (
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/jwtauth"
)

var (
	tokenAuth *jwtauth.JWTAuth
)

func Verify() func(http.Handler) http.Handler {
	return jwtauth.Verifier(tokenAuth)
}

func SetupJWTAuth(secret string) {
	parser := new(jwt.Parser)
	parser.UseJSONNumber = true
	tokenAuth = jwtauth.NewWithParser("HS256", parser, []byte(secret), nil)
}

func Encode(claims jwtauth.Claims) (*jwt.Token, error) {
	jwtToken, _, err := tokenAuth.Encode(claims)
	return jwtToken, err
}

func Decode(tok string) (*jwt.Token, error) {
	return tokenAuth.Decode(tok)
}
