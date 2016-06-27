package data

import (
	"upper.io/bond"
	"upper.io/db"
)

type PlaceStore struct {
	bond.Store
}

func (store PlaceStore) FindByID(ID int64) (*Place, error) {
	return store.FindOne(db.Cond{"id": ID})
}

func (store PlaceStore) FindOne(cond db.Cond) (*Place, error) {
	var place *Place
	if err := store.Find(cond).One(&place); err != nil {
		return nil, err
	}
	return place, nil
}

type NeighborhoodStore struct {
	bond.Store
}
