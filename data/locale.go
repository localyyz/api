package data

import (
	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
	"upper.io/bond"
	"upper.io/db.v3"
)

type Locale struct {
	ID          int64  `db:"id,pk,omitempty" json:"id,omitempty"`
	Name        string `db:"name" json:"name"`
	Shorthand   string `db:"shorthand" json:"shorthand"`
	Description string `db:"description" json:"description"`
}

type LocaleStore struct {
	bond.Store
}

var (
	EnabledLocales = db.Cond{"shorthand": []string{
		//"king-west",
		//"queen-west",
		//"distillery-district",
		//"kensington-market",
		//"st-lawrence-market",
		"west-queen-west",
		"yorkville",
	}}
)

func (n *Locale) CollectionName() string {
	return `locales`
}

func (store LocaleStore) FromLatLng(lat, lng float64) (*Locale, error) {
	// find locale from latlng
	latlng := s2.LatLngFromDegrees(lat, lng)
	origin := s2.CellIDFromLatLng(latlng).Parent(15) // 16 for more detail?
	// Find the reach of cells
	cond := db.Cond{
		"cell_id >=": int(origin.RangeMin()),
		"cell_id <=": int(origin.RangeMax()),
	}
	cells, err := DB.Cell.FindAll(cond)
	if err != nil {
		return nil, err
	}

	// Find the minimum distance cell
	min := s1.InfAngle()
	var localeID int64
	for _, c := range cells {
		cell := s2.CellID(c.CellID)
		d := latlng.Distance(cell.LatLng())
		if d < min {
			min = d
			localeID = c.LocaleID
		}
	}

	return store.FindByID(localeID)
}

func (store LocaleStore) FindByName(name string) (*Locale, error) {
	return store.FindOne(db.Cond{"name": name})
}

func (store LocaleStore) FindByID(localeID int64) (*Locale, error) {
	return store.FindOne(db.Cond{"id": localeID})
}

func (store LocaleStore) FindOne(cond db.Cond) (*Locale, error) {
	var locale *Locale
	err := DB.Locale.Find(cond).One(&locale)
	if err != nil {
		return nil, err
	}
	return locale, nil
}
