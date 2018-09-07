package data

import (
	"upper.io/bond"
	db "upper.io/db.v3"
)

// category white list
type Whitelist struct {
	CategoryID *int64        `db:"category_id,omitempty"`
	Value      string        `db:"value"`
	Gender     ProductGender `db:"gender"`
	Weight     int32         `db:"weight"`

	// special whitelist tags
	IsSpecial bool `db:"-"`
	IsIgnore  bool `db:"-"`

	// legacy columns
	Type ProductCategoryType `db:"type"`
}

type WhitelistStore struct {
	bond.Store
}

func (*Whitelist) CollectionName() string {
	return `whitelist`
}

func (store WhitelistStore) FindAll(cond db.Cond) ([]*Whitelist, error) {
	var list []*Whitelist
	if err := store.Find(cond).All(&list); err != nil {
		return nil, err
	}
	return list, nil
}
