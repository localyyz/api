package data

import (
	"fmt"
	"strings"

	"upper.io/bond"
	"upper.io/db.v3"
	"upper.io/db.v3/postgresql"
)

var (
	DB *Database
)

type Database struct {
	bond.Session

	User        UserStore
	UserAddress UserAddressStore
	Following   FollowingStore

	Place         PlaceStore
	PlaceDiscount PlaceDiscountStore
	Locale        LocaleStore
	Cell          CellStore

	ShopifyCred     ShopifyCredStore
	Product         ProductStore
	ProductTag      ProductTagStore
	ProductVariant  ProductVariantStore
	ProductCategory ProductCategoryStore
	Share           ShareStore
	Webhook         WebhookStore

	Cart     CartStore
	CartItem CartItemStore
}

type DBConf struct {
	Database        string   `toml:"database"`
	Hosts           []string `toml:"hosts"`
	Username        string   `toml:"username"`
	Password        string   `toml:"password"`
	DebugQueries    bool     `toml:"debug_queries"`
	ApplicationName string   `toml:"application_name"`
}

// String implements db.ConnectionURL
func (cf *DBConf) String() string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s",
		cf.Username, cf.Password, strings.Join(cf.Hosts, ","), cf.Database)
}

func NewDBSession(conf *DBConf) (*Database, error) {
	if conf.DebugQueries {
		db.DefaultSettings.SetLogging(true)
	}

	var (
		connURL postgresql.ConnectionURL
		err     error
	)

	connURL, err = postgresql.ParseURL(conf.String())
	if err != nil {
		return nil, err
	}
	// extra options
	connURL.Options = map[string]string{
		"application_name": conf.ApplicationName,
	}

	db := &Database{}
	db.Session, err = bond.Open(postgresql.Adapter, connURL)
	if err != nil {
		return nil, err
	}
	db.User = UserStore{db.Store(&User{})}
	db.UserAddress = UserAddressStore{db.Store(&UserAddress{})}
	db.Following = FollowingStore{db.Store(&Following{})}

	db.Place = PlaceStore{db.Store(&Place{})}
	db.PlaceDiscount = PlaceDiscountStore{db.Store(&PlaceDiscount{})}
	db.Locale = LocaleStore{db.Store(&Locale{})}
	db.Cell = CellStore{db.Store(&Cell{})}

	db.ShopifyCred = ShopifyCredStore{db.Store(&ShopifyCred{})}
	db.Product = ProductStore{db.Store(&Product{})}
	db.ProductTag = ProductTagStore{db.Store(&ProductTag{})}
	db.ProductVariant = ProductVariantStore{db.Store(&ProductVariant{})}
	db.ProductCategory = ProductCategoryStore{db.Store(&ProductCategory{})}
	db.Share = ShareStore{db.Store(&Share{})}
	db.Webhook = WebhookStore{db.Store(&Webhook{})}

	db.Cart = CartStore{db.Store(&Cart{})}
	db.CartItem = CartItemStore{db.Store(&CartItem{})}

	DB = db
	return db, nil
}
