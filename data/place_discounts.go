package data

import (
	"time"

	"upper.io/bond"
	db "upper.io/db.v3"
)

type PlaceDiscount struct {
	ID         int64      `db:"id,pk,omitempty" json:"id,omitempty"`
	PlaceID    int64      `db:"place_id" json:"placeID"`
	Code       string     `db:"code" json:"code"`
	ExternalID int64      `db:"external_id" json:"externalID"`
	CreatedAt  *time.Time `db:"created_at,omitempty" json:"createdAt"`
}

type PlaceDiscountStore struct {
	bond.Store
}

func (b *PlaceDiscount) CollectionName() string {
	return `place_discounts`
}

func (store PlaceDiscountStore) FindByPlaceID(placeID int64) (*PlaceDiscount, error) {
	return store.FindOne(db.Cond{"place_id": placeID})
}

func (store PlaceDiscountStore) FindOne(cond db.Cond) (*PlaceDiscount, error) {
	var discount *PlaceDiscount
	if err := store.Find(cond).One(&discount); err != nil {
		return nil, err
	}
	return discount, nil
}
