package data

import (
	"upper.io/bond"
	"upper.io/db.v3"
)

type FavouritePlace struct {
	UserID  int64 `db:"user_id" json:"userID"`
	PlaceID int64 `db:"place_id" json:"placeID"`
}

type FavouritePlaceStore struct {
	bond.Store
}

func (store *FavouritePlace) CollectionName() string {
	return `favourite_places`
}

func (store FavouritePlaceStore) FindAll(cond db.Cond) ([]*FavouritePlace, error) {
	var list []*FavouritePlace
	if err := store.Find(cond).All(&list); err != nil {
		return nil, err
	}
	return list, nil
}

func (store FavouritePlaceStore) FindByUserID(userID int64) ([]*FavouritePlace, error) {
	return store.FindAll(db.Cond{"user_id": userID})
}

func (store FavouritePlaceStore) FindByUserIDAndPlaceID(userID, placeID int64) (*FavouritePlace, error) {
	var favPlace *FavouritePlace
	err := store.Find(db.Cond{"user_id": userID, "place_id": placeID}).One(&favPlace)
	if err != nil {
		return nil, err
	}
	return favPlace, nil
}
