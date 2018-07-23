package data

import (
	"time"

	"upper.io/bond"
	db "upper.io/db.v3"
)

type UserDeal struct {
	UserID  int64            `db:"user_id"`
	DealID  int64            `db:"deal_id"`
	Status  CollectionStatus `db:"status"`
	StartAt time.Time        `db:"start_at"`
	EndAt   time.Time        `db:"end_at"`
}

type UserDealStore struct {
	bond.Store
}

func (d *UserDeal) CollectionName() string {
	return `user_deals`
}

func (store UserDealStore) FindAll(cond db.Cond) ([]*UserDeal, error) {
	var userDeals []*UserDeal
	if err := store.Find(cond).All(&userDeals); err != nil {
		return nil, err
	}
	return userDeals, nil
}

func (store UserDealStore) FindOne(cond db.Cond) (*UserDeal, error) {
	var userDeal *UserDeal
	if err := store.Find(cond).One(&userDeal); err != nil {
		return nil, err
	}
	return userDeal, nil
}
