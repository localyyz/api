package data

import (
	"time"

	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"upper.io/bond"
	db "upper.io/db.v3"
	"upper.io/db.v3/postgresql"
)

type Checkout struct {
	ID      int64  `db:"id,pk,omitempty" json:"id"`
	CartID  *int64 `db:"cart_id,omitempty" json:"cartId,omitempty"`
	UserID  int64  `db:"user_id" json:"userId"`
	PlaceID int64  `db:"place_id" json:"placeId"`

	Status CheckoutStatus `db:"status" json:"status"`

	Token  string `db:"token" json:"-"`
	Name   string `db:"name" json:"-"`
	WebURL string `db:"web_url" json:"-"`

	PaymentAccountID string `db:"payment_account_id" json:"-"`
	CustomerID       int64  `db:"customer_id" json:"-"`
	SuccessPaymentID int64  `db:"payment_id" json:"-"`

	Currency        string                   `db:"currency" json:"currency,omitempty"`
	DiscountCode    string                   `db:"discount_code" json:"discountCode"`
	AppliedDiscount *CheckoutAppliedDiscount `db:"applied_discount,omitempty" json:"appliedDiscount,omitempty"`

	// selected shipping method
	ShippingLine  *CheckoutShippingLine `db:"shipping_line,omitempty" json:"shippingLine,omitempty"`
	TotalShipping float64               `db:"total_shipping" json:"totalShipping,omitempty"`

	// applied tax lines
	TaxLines      []*CheckoutTaxLine `db:"tax_lines,omitempty" json:"taxLines,omitempty"`
	TotalTax      float64            `db:"total_tax" json:"totalTax,omitempty"`
	TaxesIncluded bool               `db:"taxes_included" json:"taxesIncluded,omitempty"`

	SubtotalPrice float64 `db:"subtotal_price" json:"subtotalPrice,omitempty"`
	TotalPrice    float64 `db:"total_price" json:"totalPrice,omitempty"`
	PaymentDue    string  `db:"payment_due" json:"paymentDue,omitempty"`

	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt,omitempty"`
	UpdatedAt *time.Time `db:"updated_at,omitempty" json:"updatedAt,omitempty"`
}

type CheckoutStatus uint32

type CheckoutAppliedDiscount struct {
	*shopify.AppliedDiscount
	*postgresql.JSONBConverter
}

type CheckoutShippingLine struct {
	*shopify.ShippingLine
	*postgresql.JSONBConverter
}

type CheckoutTaxLine struct {
	*shopify.TaxLine
	*postgresql.JSONBConverter
}

const (
	_                            CheckoutStatus = iota // 0
	CheckoutStatusPending                              // 1
	CheckoutStatusPaymentFailed                        // 2
	CheckoutStatusPaymentSuccess                       // 3
)

type CheckoutStore struct {
	bond.Store
}

func (c *Checkout) CollectionName() string {
	return `checkouts`
}

func (store CheckoutStore) FindByToken(token string) (*Checkout, error) {
	return store.FindOne(db.Cond{"token": token})
}

func (store CheckoutStore) FindAllByCartID(cartID int64) ([]*Checkout, error) {
	return store.FindAll(db.Cond{"cart_id": cartID})
}

func (store CheckoutStore) FindOne(cond db.Cond) (*Checkout, error) {
	var checkout *Checkout
	if err := store.Find(cond).One(&checkout); err != nil {
		return nil, err
	}
	return checkout, nil
}

func (store CheckoutStore) FindAll(cond db.Cond) ([]*Checkout, error) {
	var checkouts []*Checkout
	if err := store.Find(cond).All(&checkouts); err != nil {
		return nil, err
	}
	return checkouts, nil
}

func (c *Checkout) BeforeUpdate(bond.Session) error {
	c.UpdatedAt = GetTimeUTCPointer()

	return nil
}

func (c *Checkout) BeforeCreate(sess bond.Session) error {
	if err := c.BeforeUpdate(sess); err != nil {
		return err
	}

	c.UpdatedAt = nil
	c.CreatedAt = GetTimeUTCPointer()

	return nil
}
