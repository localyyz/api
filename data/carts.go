package data

import (
	"time"

	"upper.io/bond"
)

type Cart struct {
	ID     int64      `db:"id,pk,omitempty" json:"id"`
	UserID int64      `db:"user_id" json:"userId"`
	Status CartStatus `db:"status" json:"status"`

	Etc CartEtc `db:"etc" json:"etc"`

	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt"`
	UpdatedAt *time.Time `db:"updated_at,omitempty" json:"updatedAt"`
	DeletedAt *time.Time `db:"deleted_at,omitempty" json:"deletedAt"`
}

type CartStatus uint
type CartEtc struct {
	Shopify map[int64]string `json:"shopify"`
}

type CartStore struct {
	bond.Store
}

func (c *Cart) CollectionName() string {
	return `carts`
}
