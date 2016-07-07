package data

import (
	"time"

	"upper.io/bond"
)

type Promo struct {
	ID         int64      `db:"id,pk,omitempty" json:"id,omitempty"`
	PlaceID    int64      `db:"place_id" json:"placeId"`
	Multiplier int32      `db:"multiplier" json:"multiplier"`
	StartAt    time.Time  `db:"start_at" json:"startAt"`
	EndAt      time.Time  `db:"end_at" json:"endAt"`
	CreatedAt  *time.Time `db:"created_at,omitempty" json:"createdAt"`
}

type PromoStore struct {
	bond.Store
}

func (p *Promo) CollectionName() string {
	return `promos`
}
