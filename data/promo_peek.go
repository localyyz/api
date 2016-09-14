package data

import (
	"time"

	"upper.io/bond"
)

// PromoPeek keeps track of peaked promos by users
type PromoPeek struct {
	ID      int64 `db:"id,pk,omitempty" json:"id,omitempty"`
	PromoID int64 `db:"promo_id" json:"promoId"`
	UserID  int64 `db:"user_id" json:"userId"`

	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt"`
}

type PromoPeekStore struct {
	bond.Store
}

func (p *PromoPeek) CollectionName() string {
	return `promo_peeks`
}
