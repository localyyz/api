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

	Deal        DealStore
	DealProduct DealProductStore

	BillingPlan BillingPlanStore

	Place            PlaceStore
	PlaceBilling     PlaceBillingStore
	PlaceCharge      PlaceChargeStore
	PlaceDiscount    PlaceDiscountStore
	PriorityMerchant PriorityMerchantStore

	MerchantApproval MerchantApprovalStore

	Collection        CollectionStore
	CollectionProduct CollectionProductStore

	Product        ProductStore
	ProductVariant ProductVariantStore
	ProductImage   ProductImageStore
	VariantImage   VariantImageStore

	ShopifyCred    ShopifyCredStore
	Category       CategoryStore
	Blacklist      BlacklistStore
	FeatureProduct FeatureProductStore
	Webhook        WebhookStore

	Cart     CartStore
	CartItem CartItemStore
	Checkout CheckoutStore

	SearchWord SearchWordStore

	FavouriteProduct FavouriteProductStore
}

type DBConf struct {
	Database        string   `toml:"database"`
	Hosts           []string `toml:"hosts"`
	Username        string   `toml:"username"`
	Password        string   `toml:"password"`
	DebugQueries    bool     `toml:"debug_queries"`
	ApplicationName string   `toml:"application_name"`
	MaxConnection   int      `tomp:"max_connection"`
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

	db.Deal = DealStore{db.Store(&Deal{})}
	db.DealProduct = DealProductStore{db.Store(&DealProduct{})}

	db.BillingPlan = BillingPlanStore{db.Store(&BillingPlan{})}

	db.Place = PlaceStore{db.Store(&Place{})}
	db.PlaceBilling = PlaceBillingStore{db.Store(&PlaceBilling{})}
	db.PlaceCharge = PlaceChargeStore{db.Store(&PlaceCharge{})}
	db.PlaceDiscount = PlaceDiscountStore{db.Store(&PlaceDiscount{})}
	db.PriorityMerchant = PriorityMerchantStore{db.Store(&PriorityMerchant{})}

	db.MerchantApproval = MerchantApprovalStore{db.Store(&MerchantApproval{})}

	db.Collection = CollectionStore{db.Store(&Collection{})}
	db.CollectionProduct = CollectionProductStore{db.Store(&CollectionProduct{})}

	db.Product = ProductStore{db.Store(&Product{})}
	db.ProductVariant = ProductVariantStore{db.Store(&ProductVariant{})}
	db.ProductImage = ProductImageStore{db.Store(&ProductImage{})}
	db.VariantImage = VariantImageStore{db.Store(&VariantImage{})}

	db.ShopifyCred = ShopifyCredStore{db.Store(&ShopifyCred{})}
	db.Category = CategoryStore{db.Store(&Category{})}
	db.Blacklist = BlacklistStore{db.Store(&Blacklist{})}
	db.FeatureProduct = FeatureProductStore{db.Store(&FeatureProduct{})}
	db.Webhook = WebhookStore{db.Store(&Webhook{})}

	db.Cart = CartStore{db.Store(&Cart{})}
	db.CartItem = CartItemStore{db.Store(&CartItem{})}
	db.Checkout = CheckoutStore{db.Store(&Checkout{})}

	db.SearchWord = SearchWordStore{db.Store(&SearchWord{})}

	// set max db open connection
	db.SetMaxOpenConns(conf.MaxConnection)

	db.FavouriteProduct = FavouriteProductStore{db.Store(&FavouriteProduct{})}

	DB = db
	return db, nil
}
