package data

import (
	"github.com/golang/geo/s2"
	"github.com/upper/bond"
	"upper.io/db.v2"
)

// When a (lng, lat) pair is received, process it using google's s2 lib
//  and return a lv15 cellid equivalent. It is possible then to categorise those
//  cellids into different locales
// We can then build a database of cell ids
// good Blog post on s2 geometry
//	http://blog.christianperone.com/2015/08/googles-s2-geometry-on-the-sphere-cells-and-hilbert-curve/

type Cell struct {
	ID       int64 `db:"id,pk,omitempty" json:"id"`
	CellID   int64 `db:"cell_id" json:"cell_id"`
	LocaleID int64 `db:"locale_id" json:"locale_id"`
	//Level   int32 `db:"level" json:"level"` // quick lookup of cell level
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

func (store CellStore) FindByLatLng(lat, lng float64) (*Cell, error) {
	origin := s2.CellIDFromLatLng(s2.LatLngFromDegrees(lat, lng)).Parent(cellIDLevel)
	return store.FindOne(db.Cond{"cell_id": int64(origin)})
}

func (store CellStore) FindNeighbourByLatLng(lat, lng float64) ([]*Cell, error) {
	origin := s2.CellIDFromLatLng(s2.LatLngFromDegrees(lat, lng)).Parent(cellIDLevel)
	return store.FindNeighbourByCellID(int64(origin))
}

func (store CellStore) FindNeighbourByCellID(cellID int64) ([]*Cell, error) {
	origin := s2.CellID(cellID)
	neighbours := make([]int64, 4)
	for i, n := range origin.EdgeNeighbors() {
		neighbours[i] = int64(n)
	}
	return store.FindAll(db.Cond{"cell_id": neighbours})
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
