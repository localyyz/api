package data

import (
	"fmt"
	"strings"
	"time"

	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"upper.io/bond"
	db "upper.io/db.v3"
	"upper.io/db.v3/lib/sqlbuilder"
	"upper.io/db.v3/postgresql"
)

type Cart struct {
	ID     int64      `db:"id,pk,omitempty" json:"id"`
	UserID int64      `db:"user_id" json:"userId"`
	Status CartStatus `db:"status" json:"status"`

	IsExpress bool `db:"is_express" json:"isExpress"`

	Etc             CartEtc      `db:"etc" json:"etc"`
	ShippingAddress *CartAddress `db:"shipping_address,omitempty" json:"shippingAddress"`
	BillingAddress  *CartAddress `db:"billing_address,omitempty" json:"billingAddress"`
	Email           string       `db:"email,omitempty" json:"email"`

	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt"`
	UpdatedAt *time.Time `db:"updated_at,omitempty" json:"updatedAt"`
	DeletedAt *time.Time `db:"deleted_at,omitempty" json:"deletedAt"`
}

type CartStatus uint

type CartAddress struct {
	// Core data
	FirstName    string `json:"firstName,omitempty"`
	LastName     string `json:"lastName,omitempty"`
	Address      string `json:"address,omitempty"`
	AddressOpt   string `json:"addressOpt,omitempty"`
	City         string `json:"city,omitempty"`
	Country      string `json:"country,omitempty"`
	CountryCode  string `json:"countryCode,omitempty"`
	Province     string `json:"province,omitempty"`
	ProvinceCode string `json:"provinceCode,omitempty"`
	Zip          string `json:"zip,omitempty"`

	// Extra data
	Email string `json:"email,omitempty"`
	Phone string `json:"phone,omitempty"`

	// Internal ID
	ID int64 `json:"id,omitempty"`
	*postgresql.JSONBConverter
}

// normalize trims all fields
func (a *CartAddress) Normalize() {
	a.FirstName = strings.TrimSpace(a.FirstName)
	a.LastName = strings.TrimSpace(a.LastName)
	a.Address = strings.TrimSpace(a.Address)
	a.AddressOpt = strings.TrimSpace(a.AddressOpt)
	a.City = strings.TrimSpace(a.City)
	a.Country = strings.TrimSpace(a.Country)
	a.CountryCode = strings.TrimSpace(a.CountryCode)
	a.Province = strings.TrimSpace(a.Province)
	a.ProvinceCode = strings.TrimSpace(a.ProvinceCode)
	a.Zip = strings.TrimSpace(a.Zip)
	a.Email = strings.TrimSpace(a.Email)
	a.Phone = strings.TrimSpace(a.Phone)
}

type CartShippingMethod struct {
	Handle        string      `json:"handle"`
	Title         string      `json:"title"`
	Price         int64       `json:"price"`
	DeliveryRange []time.Time `json:"deliveryRange"`
}

type CartEtc struct {
	// placeID -> shopify data
	ShopifyData     map[int64]*CartShopifyData    `json:"shopifyData,omitempty"`
	ShippingMethods map[int64]*CartShippingMethod `json:"shippingMethods,omitempty"`
	ShippingAddress *CartAddress                  `json:"shippingAddress,omitempty"`
	BillingAddress  *CartAddress                  `json:"billingAddress,omitempty"`
	DiscountCode    string                        `json:"discountCode,omitempty"`
	*postgresql.JSONBConverter
}

type CartShopifyData struct {
	Token                   string `json:"token"`
	ShopifyPaymentAccountID string `json:"shopify_payments_account_id"`
	PaymentURL              string `json:"payment_url"`
	Name                    string `json:"name"`
	CustomerID              int64  `json:"customer_id"`
	WebURL                  string `json:"webUrl"`
	WebProcessingURL        string `json:"webProcessingUrl"`

	PaymentID int64 `json:"paymentId"`

	Currency   string                   `json:"currency,omitempty"`
	Discount   *shopify.AppliedDiscount `json:"discount,omitempty"`
	TotalTax   int64                    `json:"totalTax"`
	TotalPrice int64                    `json:"totalPrice"`
	PaymentDue string                   `json:"paymentDue"`

	// INTERNAL place id reference
	PlaceID int64 `json:"-"`
}

type CartStore struct {
	bond.Store
}

const (
	CartStatusUnknown         CartStatus = iota // 0
	CartStatusInProgress                        // 1
	CartStatusCheckout                          // 2
	CartStatusPaymentSuccess                    // 3
	CartStatusComplete                          // 4
	CartStatusPartialCheckout                   // 5
)

var _ interface {
	bond.HasBeforeCreate
	bond.HasBeforeUpdate
} = &Cart{}
var _ sqlbuilder.ValueWrapper = &CartEtc{}

var (
	cartStatuses = []string{
		"",
		"inprogress",
		"checkout",
		"payment_success",
		"completed",
		"partial_checkout",
	}
)

func (c *Cart) CollectionName() string {
	return `carts`
}

func (store CartStore) FindByID(ID int64) (*Cart, error) {
	return store.FindOne(db.Cond{"id": ID})
}

func (store CartStore) FindByUserID(userID int64) ([]*Cart, error) {
	return store.FindAll(db.Cond{"user_id": userID})
}

func (store CartStore) FindOne(cond db.Cond) (*Cart, error) {
	var cart *Cart
	if err := DB.Cart.Find(cond).One(&cart); err != nil {
		return nil, err
	}
	return cart, nil
}

func (store CartStore) FindAll(cond db.Cond) ([]*Cart, error) {
	var carts []*Cart
	if err := DB.Cart.Find(cond).All(&carts); err != nil {
		return nil, err
	}
	return carts, nil
}

func (c *Cart) BeforeUpdate(bond.Session) error {
	c.UpdatedAt = GetTimeUTCPointer()

	return nil
}

func (c *Cart) BeforeCreate(sess bond.Session) error {
	if err := c.BeforeUpdate(sess); err != nil {
		return err
	}

	c.UpdatedAt = nil
	c.CreatedAt = GetTimeUTCPointer()

	return nil
}

// String returns the string value of the status.
func (s CartStatus) String() string {
	return cartStatuses[s]
}

// MarshalText satisfies TextMarshaler
func (s CartStatus) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalText satisfies TextUnmarshaler
func (s *CartStatus) UnmarshalText(text []byte) error {
	enum := string(text)
	for i := 0; i < len(cartStatuses); i++ {
		if enum == cartStatuses[i] {
			*s = CartStatus(i)
			return nil
		}
	}
	return fmt.Errorf("unknown cart status %s", enum)
}
