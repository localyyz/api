package data

import (
	"time"

	"upper.io/bond"
)

type Claim struct {
	ID      int64 `db:"id,pk,omitempty" json:"id,omitempty"`
	PromoID int64 `db:"promo_id" json:"promoId"`
	PlaceID int64 `db:"place_id" json:"placeId"`
	UserID  int64 `db:"user_id" json:"userId"`

	Status ClaimStatus `db:"status" json:"status"`

	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt"`
}

type ClaimStore struct {
	bond.Store
}

type ClaimStatus uint32

const (
	_ ClaimStatus = iota
	ClaimStatusActive
	ClaimStatusComplete
	ClaimStatusExpired
)
