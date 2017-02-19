package data

import (
	"upper.io/bond"
	"upper.io/db.v3"
)

type ClaimStore struct {
	bond.Store
}

func (store ClaimStore) FindByUserID(userID int64) ([]*Claim, error) {
	return store.FindAll(db.Cond{"user_id": userID})
}

func (store ClaimStore) FindOne(cond db.Cond) (*Claim, error) {
	var claim *Claim
	if err := store.Find(cond).One(&claim); err != nil {
		return nil, err
	}
	return claim, nil
}

func (store ClaimStore) FindAll(cond db.Cond) ([]*Claim, error) {
	var claims []*Claim
	if err := store.Find(cond).All(&claims); err != nil {
		return nil, err
	}
	return claims, nil
}
