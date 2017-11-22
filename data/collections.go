package data

import (
	"time"

	"upper.io/bond"
	"upper.io/db.v3"
)

type Collection struct {
	ID          int64      `db:"id,pk,omitempty" json:"id,omitempty"`
	Name        string     `db:"name" json:"name"`
	Description string     `db:"description" json:"description"`
	ImageURL    string     `db:"image_url" json:"imageUrl"`
	Ordering    int32      `db:"ordering" json:"ordering"`
	CreatedAt   *time.Time `db:"created_at,omitempty" json:"createdAt,omitempty"`
}

type CollectionStore struct {
	bond.Store
}

func (*Collection) CollectionName() string {
	return `collections`
}

type CollectionProduct struct {
	ID           int64      `db:"id,pk,omitempty" json:"id,omitempty"`
	CollectionID int64      `db:"collection_id" json:"collection_id"`
	ProductID    int64      `db:"product_id" json:"product_id"`
	CreatedAt    *time.Time `db:"created_at,omitempty" json:"createdAt,omitempty"`
}

type CollectionProductStore struct {
	bond.Store
}

func (*CollectionProduct) CollectionName() string {
	return `collection_products`
}

func (store CollectionStore) FindByID(ID int64) (*Collection, error) {
	return store.FindOne(db.Cond{"id": ID})
}

func (store CollectionStore) FindOne(cond db.Cond) (*Collection, error) {
	var collection *Collection
	err := DB.Collection.Find(cond).One(&collection)
	if err != nil {
		return nil, err
	}
	return collection, nil
}

func (store CollectionStore) FindAll(cond db.Cond) ([]*Collection, error) {
	var collections []*Collection
	err := DB.Collection.Find(cond).All(&collections)
	if err != nil {
		return nil, err
	}
	return collections, nil
}

func (store CollectionProductStore) FindOne(cond db.Cond) (*CollectionProduct, error) {
	var collection *CollectionProduct
	err := DB.CollectionProduct.Find(cond).One(&collection)
	if err != nil {
		return nil, err
	}
	return collection, nil
}

func (store CollectionProductStore) FindAll(cond db.Cond) ([]*CollectionProduct, error) {
	var collections []*CollectionProduct
	err := DB.CollectionProduct.Find(cond).All(&collections)
	if err != nil {
		return nil, err
	}
	return collections, nil
}
