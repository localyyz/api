package data

import (
	"time"

	"upper.io/bond"
	db "upper.io/db.v3"
)

type PlaceDiscount struct {
	ID         int64      `db:"id,pk,omitempty" json:"id,omitempty"`
	PlaceID    int64      `db:"place_id" json:"placeId"`
	Code       string     `db:"code" json:"code"`
	ExternalID int64      `db:"external_id" json:"externalId"`
	CreatedAt  *time.Time `db:"created_at,omitempty" json:"createdAt,omitempty"`
}

type PlaceDiscountStore struct {
	bond.Store
}

func (p *PlaceDiscount) CollectionName() string {
	return `place_discounts`
}

func (store PlaceDiscountStore) FindByPlaceID(placeID int64) ([]*PlaceDiscount, error) {
	return store.FindAll(db.Cond{"place_id": placeID})
}

func (store PlaceDiscountStore) FindOne(cond db.Cond) (*PlaceDiscount, error) {
	var discount *PlaceDiscount
	if err := store.Find(cond).One(&discount); err != nil {
		return nil, err
	}
	return discount, nil
}

func (store PlaceDiscountStore) FindAll(cond db.Cond) ([]*PlaceDiscount, error) {
	var discounts []*PlaceDiscount
	if err := store.Find(cond).All(&discounts); err != nil {
		return nil, err
	}
	return discounts, nil
}
