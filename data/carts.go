package data

import (
	"fmt"
	"time"

	"upper.io/bond"
	db "upper.io/db.v3"
)

type Cart struct {
	ID     int64      `db:"id,pk,omitempty" json:"id"`
	UserID int64      `db:"user_id" json:"userId"`
	Status CartStatus `db:"status" json:"status"`

	Etc CartEtc `db:"etc,jsonb" json:"etc"`

	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt"`
	UpdatedAt *time.Time `db:"updated_at,omitempty" json:"updatedAt"`
	DeletedAt *time.Time `db:"deleted_at,omitempty" json:"deletedAt"`
}

type CartStatus uint

type CartAddress struct {
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	Address    string `json:"address"`
	AddressOpt string `json:"addressOpt"`
	City       string `json:"city"`
	Country    string `json:"country"`
	Province   string `json:"province"`
	Zip        string `json:"zip"`
}

type CartShippingMethod struct {
	Handle        string      `json:"handle"`
	Title         string      `json:"title"`
	Price         float64     `json:"price"`
	SubtotalPrice float64     `json:"subtotalPrice"`
	TotalTax      float64     `json:"totalTax"`
	TotalPrice    float64     `json:"totalPrice"`
	DeliveryRange []time.Time `json:"deliveryRange"`
}

type CartEtc struct {
	// placeID -> shopify data
	ShopifyData     map[int64]CartShopifyData `json:"shopifyData,omitempty"`
	ShippingAddress *CartAddress              `json:"shippingAddress,omitempty"`
	ShippingMethod  *CartShippingMethod       `json:"shippingMethod,omitempty"`
}

type CartShopifyData struct {
	Token            string `json:"token"`
	PaymentAccountID string `json:"payments_account_id"`
	Name             string `json:"name"`
	CustomerID       int64  `json:"customer_id"`
	WebURL           string `json:"webUrl"`
	WebProcessingURL string `json:"webProcessingUrl"`
	HasPayed         bool   `json:"hasPayed"`
}

type CartStore struct {
	bond.Store
}

const (
	CartStatusUnknown CartStatus = iota
	CartStatusInProgress
	CartStatusProcessing
	CartStatusComplete
)

var _ interface {
	bond.HasBeforeCreate
	bond.HasBeforeUpdate
} = &Cart{}

var (
	cartStatuses = []string{
		"",
		"inprogress",
		"processing",
		"completed",
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
