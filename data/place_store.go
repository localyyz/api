package data

import (
	"fmt"

	"upper.io/bond"
	"upper.io/db.v3"
)

type PlaceStore struct {
	bond.Store
}

func (store PlaceStore) FindByLocaleID(localeID int64) ([]*Place, error) {
	return store.FindAll(db.Cond{"locale_id": localeID})
}

func (store PlaceStore) MatchName(q string) ([]*Place, error) {
	return store.FindAll(db.Cond{"name ~*": fmt.Sprint("\\m(", q, ")")})
}

func (store PlaceStore) FindByCategory(category Category) ([]*Place, error) {
	return store.FindAll(db.Cond{"category": category})
}

func (store PlaceStore) FindAll(cond db.Cond) ([]*Place, error) {
	var places []*Place
	if err := store.Find(cond).OrderBy("name").All(&places); err != nil {
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
