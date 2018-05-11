package data

import (
	"upper.io/bond"
	db "upper.io/db.v3"
)

type Blacklist struct {
	Word string `db:"word" json:"word"`
}

type BlacklistStore struct {
	bond.Store
}

func (p *Blacklist) CollectionName() string {
	return `blacklist`
}

func (store BlacklistStore) FindAll(cond db.Cond) ([]*Blacklist, error) {
	var list []*Blacklist
	if err := store.Find(cond).All(&list); err != nil {
		return nil, err
	}
	return list, nil
}
