package data

import (
	"time"
	"upper.io/bond"
	"upper.io/db.v3"
)

type UserCollection struct {
	ID        int64      `db:"id,pk,omitempty" json:"id"`
	UserID    int64      `db:"user_id" json:"userId"`
	Title     string     `db:"title" json:"title"`
	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt"`
	UpdatedAt *time.Time `db:"updated_at,omitempty" json:"updatedAt"`
	DeletedAt *time.Time `db:"deleted_at,omitempty" json:"deletedAt"`
}

type UserCollectionStore struct {
	bond.Store
}

func (store *UserCollection) CollectionName() string {
	return `user_collections`
}

func (store UserCollectionStore) FindAll(cond db.Cond) ([]*UserCollection, error) {
	var list []*UserCollection
	if err := store.Find(cond).OrderBy("updated_at DESC").All(&list); err != nil {
		return nil, err
	}
	return list, nil
}

func (store UserCollectionStore) FindByID(userID, collectionID int64) (*UserCollection, error) {
	var userCollection *UserCollection
	err := store.Find(db.Cond{"id": collectionID, "user_id": userID, "deleted_at": db.IsNull()}).One(&userCollection)
	return userCollection, err
}

func (store UserCollectionStore) FindByUserID(userID int64) ([]*UserCollection, error) {
	return store.FindAll(db.Cond{"user_id": userID, "deleted_at": db.IsNull()})
}

type UserCollectionProduct struct {
	CollectionID int64      `db:"collection_id" json:"collectionID"`
	ProductID    int64      `db:"product_id" json:"productID"`
	CreatedAt    *time.Time `db:"created_at,omitempty" json:"createdAt"`
	DeletedAt    *time.Time `db:"deleted_at" json:"deletedAt"`
}

type UserCollectionProductStore struct {
	bond.Store
}

func (store *UserCollectionProduct) CollectionName() string {
	return `user_collection_products`
}

func (store UserCollectionProductStore) FindAll(cond db.Cond) ([]*UserCollectionProduct, error) {
	var list []*UserCollectionProduct
	if err := store.Find(cond).All(&list); err != nil {
		return nil, err
	}
	return list, nil
}

func (store UserCollectionProductStore) FindByProductID(productID int64) ([]*UserCollectionProduct, error) {
	return store.FindAll(db.Cond{"product_id": productID, "deleted_at": db.IsNull()})
}

func (store UserCollectionProductStore) FindByCollectionID(collectionID int64) ([]*UserCollectionProduct, error) {
	return store.FindAll(db.Cond{"collection_id": collectionID, "deleted_at": db.IsNull()})
}

func (store UserCollectionProductStore) FindByCollectionAndProductID(collectionID, productID int64) (*UserCollectionProduct, error) {
	var uP UserCollectionProduct
	err := store.Find(db.Cond{"product_id": productID, "collection_id": collectionID, "deleted_at": db.IsNull()}).Limit(1).One(&uP)
	return &uP, err
}
