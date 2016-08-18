package data

import (
	"github.com/golang/geo/s2"
	"upper.io/bond"
	db "upper.io/db.v2"
)

// When a (lng, lat) pair is received, process it using google's s2 lib
//  and return a lv15 cellid equivalent. It is possible then to categorise those
//  cellids into different locales
// We can then build a database of cell ids
// good Blog post on s2 geometry
//	http://blog.christianperone.com/2015/08/googles-s2-geometry-on-the-sphere-cells-and-hilbert-curve/

type Cell struct {
	ID      int64 `db:"id,pk" json:"id"`
	LocalID int64 `db:"locale_id" json:"locale_id"`
	Level   int32 `db:"level" json:"level"` // quick lookup of cell level
	// FUTURE/TODO cityID
}

type CellStore struct {
	bond.Store
}

var (
	cellIDLevel = 15
)

func (c *Cell) CollectionName() string {
	return `cells`
}

func (store CellStore) FindNearby(cellID int64) ([]*Cell, error) {
	origin := s2.CellID(cellID)
	cellIDs := []s2.CellID{origin}
	for _, cellID := range origin.EdgeNeighbors() {
		cellIDs = append(cellIDs, cellID)
	}
	return store.FindAll(db.Cond{"id": cellIDs})
}

func (store CellStore) FindByLatLng(lat, lng float64) (*Cell, error) {
	origin := s2.CellIDFromLatLng(s2.LatLngFromDegrees(lat, lng)).Parent(cellIDLevel)
	return store.FindOne(db.Cond{"id": origin})
}

func (store CellStore) FindOne(cond db.Cond) (*Cell, error) {
	var cell *Cell
	if err := store.Find(cond).One(&cell); err != nil {
		return nil, err
	}
	return cell, nil
}

func (store CellStore) FindAll(cond db.Cond) ([]*Cell, error) {
	var cells []*Cell
	if err := store.Find(cond).All(&cells); err != nil {
		return nil, err
	}
	return cells, nil
}
