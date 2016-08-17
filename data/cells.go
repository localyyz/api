package data

// When a (lng, lat) pair is received, process it using google's s2 lib
//  and return a lv15 cellid equivalent. It is possible then to categorise those
//  cellids into different locales
// We can then build a database of cell ids
// good Blog post on s2 geometry
//	http://blog.christianperone.com/2015/08/googles-s2-geometry-on-the-sphere-cells-and-hilbert-curve/

type Cell struct {
	ID      int64 `db:"id,pk" json:"id"`
	LocalID int64 `db:"locale_id" json:"locale_id"`
	// FUTURE/TODO cityID
}

var (
	cellIDLevel         = 15
	earthRadiusInMeters = 6378100
)

func (c *Cell) CollectionName() string {
	return `cells`
}
