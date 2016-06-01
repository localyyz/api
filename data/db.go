package data

import (
	"fmt"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/goware/jwtauth"

	"upper.io/bond"
	"upper.io/db/postgresql"
)

var (
	DB        *Database
	tokenAuth *jwtauth.JwtAuth
)

type Database struct {
	bond.Session

	Account AccountStore
}

type DBConf struct {
	Database string   `toml:"database"`
	Hosts    []string `toml:"hosts"`
	Username string   `toml:"username"`
	Password string   `toml:"password"`
}

// ConnectionUrl implements db.ConnectionURL
func (cf *DBConf) ConnectionUrl() string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s",
		cf.Username, cf.Password, strings.Join(cf.Hosts, ","), cf.Database)
}

func NewDBSession(conf DBConf) error {

	connUrl, err := postgresql.ParseURL(conf.ConnectionUrl())
	if err != nil {
		return err
	}

	db := &Database{}
	db.Session, err = bond.Open(postgresql.Adapter, connUrl)
	if err != nil {
		return err
	}
	db.Account = AccountStore{db.Store(&Account{})}

	DB = db
	return nil
}

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
