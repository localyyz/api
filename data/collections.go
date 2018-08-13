package data

import (
	"time"

	"upper.io/bond"
	"upper.io/db.v3"
	"upper.io/db.v3/postgresql"
)

type Collection struct {
	ID          int64         `db:"id,pk,omitempty" json:"id,omitempty"`
	Name        string        `db:"name" json:"name"`
	Description string        `db:"description" json:"description"`
	ImageURL    string        `db:"image_url" json:"imageUrl"`
	ImageWidth  int64         `db:"image_width" json:"imageWidth"`
	ImageHeight int64         `db:"image_height" json:"imageHeight"`
	Gender      ProductGender `db:"gender" json:"gender"`
	Featured    bool          `db:"featured" json:"featured"`

	PlaceIDs   *postgresql.Int64Array  `db:"place_ids" json:"-"`
	Categories *postgresql.StringArray `db:"categories" json:"-"`

	Ordering   int32      `db:"ordering" json:"ordering"`
	CreatedAt  *time.Time `db:"created_at,omitempty" json:"createdAt,omitempty"`
	ExternalID *int64     `db:"external_id,omitempty" json:"-"`

	MerchantID int64 `db:"merchant_id" json:"-"`
}

type CollectionStore struct {
	bond.Store
}

func (*Collection) CollectionName() string {
	return `collections`
}

type CollectionProduct struct {
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

func (store CollectionProductStore) FindByCollectionID(collectionID int64) ([]*CollectionProduct, error) {
	return DB.CollectionProduct.FindAll(db.Cond{"collection_id": collectionID})
}

func (store CollectionProductStore) FindAll(cond db.Cond) ([]*CollectionProduct, error) {
	var collections []*CollectionProduct
	err := DB.CollectionProduct.Find(cond).All(&collections)
	if err != nil {
		return nil, err
	}
	return collections, nil
}
