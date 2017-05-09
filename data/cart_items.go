package data

import (
	"time"

	"upper.io/bond"
)

type CartItem struct {
	ID     int64 `db:"id,pk,omitempty" json:"id"`
	UserID int64 `db:"user_id" json:"userId"`
	CartID int64 `db:"cart_id" json:"cart_id"`

	ProductID int64 `db:"product_id" json:"product_id"`
	VariantID int64 `db:"variant_id" json:"variant_id"`

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
