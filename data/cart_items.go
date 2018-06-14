package data

import (
	"time"

	"upper.io/bond"
	db "upper.io/db.v3"
)

type CartItem struct {
	ID         int64  `db:"id,pk,omitempty" json:"id"`
	CartID     int64  `db:"cart_id" json:"cartId"`
	CheckoutID *int64 `db:"checkout_id,omitempty" json:"checkoutId"`

	ProductID int64 `db:"product_id" json:"productId"`
	VariantID int64 `db:"variant_id" json:"variantId"`
	PlaceID   int64 `db:"place_id" json:"placeId"`

	Quantity uint32 `db:"quantity" json:"quantity"`

	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt"`
	UpdatedAt *time.Time `db:"updated_at,omitempty" json:"updatedAt"`
	DeletedAt *time.Time `db:"deleted_at,omitempty" json:"deletedAt"`
}

type CartItemStore struct {
	bond.Store
}

func (c *CartItem) CollectionName() string {
	return `cart_items`
}

var _ interface {
	bond.HasBeforeCreate
	bond.HasBeforeUpdate
} = &CartItem{}

func (store CartItemStore) FindByID(ID int64) (*CartItem, error) {
	return store.FindOne(db.Cond{"id": ID})
}

func (store CartItemStore) FindByCartID(cartID int64) ([]*CartItem, error) {
	return store.FindAll(db.Cond{"cart_id": cartID})
}

func (store CartItemStore) FindByCheckoutID(checkoutID int64) ([]*CartItem, error) {
	return store.FindAll(db.Cond{"checkout_id": checkoutID})
}

func (store CartItemStore) FindOne(cond db.Cond) (*CartItem, error) {
	var cartItem *CartItem
	if err := DB.CartItem.Find(cond).One(&cartItem); err != nil {
		return nil, err
	}
	return cartItem, nil
}

func (store CartItemStore) FindAll(cond db.Cond) ([]*CartItem, error) {
	var cartItems []*CartItem
	if err := DB.CartItem.Find(cond).All(&cartItems); err != nil {
		return nil, err
	}
	return cartItems, nil
}

func (c *CartItem) BeforeUpdate(bond.Session) error {
	c.UpdatedAt = GetTimeUTCPointer()

	return nil
}

func (c *CartItem) BeforeCreate(sess bond.Session) error {
	if err := c.BeforeUpdate(sess); err != nil {
		return err
	}

	c.UpdatedAt = nil
	c.CreatedAt = GetTimeUTCPointer()

	return nil
}
