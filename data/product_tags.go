package data

import (
	"time"

	"upper.io/bond"
	db "upper.io/db.v3"
)

type ProductTag struct {
	ID        int64  `db:"id,pk,omitempty"`
	ProductID int64  `db:"product_id"`
	Value     string `db:"value"`

	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt"`
}

type ProductTagStore struct {
	bond.Store
}

func (t *ProductTag) CollectionName() string {
	return `product_tags`
}

func (store ProductTagStore) FindByProduct(productID int64) ([]*ProductTag, error) {
	return store.FindAll(db.Cond{"product_id": productID})
}

func (store ProductTagStore) FindAll(cond db.Cond) ([]*ProductTag, error) {
	var tags []*ProductTag
	if err := store.Find(cond).All(&tags); err != nil {
		return nil, err
	}
	return tags, nil
}
