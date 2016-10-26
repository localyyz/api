package data

import (
	"fmt"
	"strings"

	"github.com/upper/bond"

	db "upper.io/db.v2"
	"upper.io/db.v2/postgresql"
)

var (
	DB *Database
)

type Database struct {
	bond.Session

	User UserStore

	Place  PlaceStore
	Locale LocaleStore
	Cell   CellStore

	Promo PromoStore
	Claim ClaimStore
}

type DBConf struct {
	Database     string   `toml:"database"`
	Hosts        []string `toml:"hosts"`
	Username     string   `toml:"username"`
	Password     string   `toml:"password"`
	DebugQueries bool     `toml:"debug_queries"`
}

// ConnectionUrl implements db.ConnectionURL
func (cf *DBConf) ConnectionUrl() string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s",
		cf.Username, cf.Password, strings.Join(cf.Hosts, ","), cf.Database)
}

func NewDBSession(conf DBConf) error {
	if conf.DebugQueries {
		db.Conf.SetLogging(true)
	}

	connUrl, err := postgresql.ParseURL(conf.ConnectionUrl())
	if err != nil {
		return err
	}

	db := &Database{}
	db.Session, err = bond.Open(postgresql.Adapter, connUrl)
	if err != nil {
		return err
	}
	db.User = UserStore{db.Store(&User{})}

	db.Place = PlaceStore{db.Store(&Place{})}
	db.Locale = LocaleStore{db.Store(&Locale{})}
	db.Cell = CellStore{db.Store(&Cell{})}

	db.Promo = PromoStore{db.Store(&Promo{})}
	db.Claim = ClaimStore{db.Store(&Claim{})}

	DB = db
	return nil
}
