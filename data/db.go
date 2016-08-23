package data

import (
	"fmt"
	"strings"

	"upper.io/bond"
	"upper.io/db.v2/postgresql"
)

var (
	DB *Database
)

type Database struct {
	bond.Session

	User      UserStore
	Post      PostStore
	Like      LikeStore
	Comment   CommentStore
	UserPoint UserPointStore

	Place  PlaceStore
	Locale LocaleStore
	Cell   CellStore

	Promo PromoStore
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
	db.User = UserStore{db.Store(&User{})}
	db.Post = PostStore{db.Store(&Post{})}
	db.Like = LikeStore{db.Store(&Like{})}
	db.Comment = CommentStore{db.Store(&Comment{})}
	db.UserPoint = UserPointStore{db.Store(&UserPoint{})}

	db.Place = PlaceStore{db.Store(&Place{})}
	db.Locale = LocaleStore{db.Store(&Locale{})}
	db.Cell = CellStore{db.Store(&Cell{})}

	db.Promo = PromoStore{db.Store(&Promo{})}

	DB = db
	return nil
}
