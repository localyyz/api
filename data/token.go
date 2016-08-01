package data

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/goware/jwtauth"
)

var (
	tokenAuth *jwtauth.JwtAuth
)

func SetupJWTAuth(secret string) {
	parser := new(jwt.Parser)
	parser.UseJSONNumber = true
	tokenAuth = jwtauth.NewWithParser("HS256", parser, []byte(secret), nil)
}

func GenerateToken(userID int64) (*jwt.Token, error) {
	claims := jwtauth.Claims{"user_id": userID}

	jwtToken, _, err := tokenAuth.Encode(claims)
	return jwtToken, err
}

func DecodeToken(tok string) (*jwt.Token, error) {
	return tokenAuth.Decode(tok)
}
