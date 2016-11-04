package data

import (
	"fmt"

	"github.com/upper/bond"
	"upper.io/db.v2"
)

type PlaceStore struct {
	bond.Store
}

func (store PlaceStore) FindByLocaleID(localeID int64) ([]*Place, error) {
	return store.FindAll(db.Cond{"locale_id": localeID})
}

func (store PlaceStore) FindLikeName(q string) ([]*Place, error) {
	return store.FindAll(db.Cond{"name ILIKE": fmt.Sprint("%", q, "%")})
}

func (store PlaceStore) FindAll(cond db.Cond) ([]*Place, error) {
	var places []*Place
	if err := store.Find(cond).All(&places); err != nil {
		return nil, err
	}
	return places, nil
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
