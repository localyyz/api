package data

import (
	"time"

	"upper.io/bond"
	"upper.io/db.v3"
)

type FavouriteProduct struct {
	UserID    int64      `db:"user_id" json:"userID"`
	ProductID int64      `db:"product_id" json:"productID"`
	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt"`
}

type FavouriteProductStore struct {
	bond.Store
}

func (store *FavouriteProduct) CollectionName() string {
	return `favourite_products`
}

func (store FavouriteProductStore) FindAll(cond db.Cond) ([]*FavouriteProduct, error) {
	var list []*FavouriteProduct
	if err := store.Find(cond).All(&list); err != nil {
		return nil, err
	}
	return list, nil
}

func (store FavouriteProductStore) FindByUserID(userID int64) ([]*FavouriteProduct, error) {
	return store.FindAll(db.Cond{"user_id": userID})
}

func (store FavouriteProductStore) GetProductIDsFromUser(userID int64) ([]int64, error) {
	favs, err := store.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	productIDs := make([]int64, len(favs))
	for i, f := range favs {
		productIDs[i] = f.ProductID
	}
	return productIDs, nil
}

func (store FavouriteProductStore) FindByProductID(productID int64) ([]*FavouriteProduct, error) {
	return store.FindAll(db.Cond{"product_id": productID})
}

func (store FavouriteProductStore) FindByUserIDAndProductID(userID, productID int64) (*FavouriteProduct, error) {
	var favProduct *FavouriteProduct
	err := store.Find(db.Cond{"user_id": userID, "product_id": productID}).One(&favProduct)
	if err != nil {
		return nil, err
	}
	return favProduct, nil
}
