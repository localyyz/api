package data

import (
	"upper.io/bond"
	db "upper.io/db.v3"
)

type ProductQueue struct {
	Id       int64  `db:"id" json:"id"`
	ImageURL string `db:"image_url" json:"image_url"`
	Tags     string `db:"tags" json:"tags"`
}

type ProductQueueStore struct {
	bond.Store
}

func (p *ProductQueue) CollectionName() string {
	return `product_queue`
}

func (store ProductQueueStore) FindAll(cond db.Cond) ([]*ProductQueue, error) {
	var list []*ProductQueue
	if err := store.Find(cond).All(&list); err != nil {
		return nil, err
	}
	return list, nil
}
