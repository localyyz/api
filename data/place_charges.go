package data

import (
	"time"

	"upper.io/bond"
	db "upper.io/db.v3"
)

type PlaceCharge struct {
	ID         int64      `db:"id,pk,omitempty" json:"id,omitempty"`
	PlaceID    int64      `db:"place_id" json:"placeID"`
	ExternalID int64      `db:"external_id" json:"externalID"`
	ChargeType ChargeType `db:"charge_type" json:"chargeType"`
	Amount     float64    `db:"amount" json:"amount"`
	CreatedAt  *time.Time `db:"created_at,omitempty" json:"createdAt"`
	ExpireAt   *time.Time `db:"expire_at,omitempty" json:"expireAt"`
}

type ChargeType uint32

type PlaceChargeStore struct {
	bond.Store
}

const (
	_ ChargeType = iota
	ChargeTypeSubscription
	ChargeTypeTransactionFee
	ChargeTypeCommissionFee
)

var (
	chargeTypes = []string{
		"-",
		"subscription",
		"transaction",
		"commission",
	}
)

func (b *PlaceCharge) CollectionName() string {
	return `place_charges`
}

func (store PlaceChargeStore) FindActiveSubscriptionCharge(placeID int64) (*PlaceCharge, error) {
	return store.FindOne(
		db.Cond{
			"place_id":    placeID,
			"charge_type": ChargeTypeSubscription,
			"expire_at":   db.Lt(time.Now()),
		},
	)
}

func (store PlaceChargeStore) FindAll(cond db.Cond) ([]*PlaceCharge, error) {
	var charges []*PlaceCharge
	if err := store.Find(cond).All(&charges); err != nil {
		return nil, err
	}
	return charges, nil
}

func (store PlaceChargeStore) FindOne(cond db.Cond) (*PlaceCharge, error) {
	var charge *PlaceCharge
	if err := store.Find(cond).One(&charge); err != nil {
		return nil, err
	}
	return charge, nil
}
