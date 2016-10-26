package locale

import (
	"fmt"

	"github.com/golang/geo/s2"
)

type Bound struct {
	North float64 `json:"north"`
	South float64 `json:"south"`
	East  float64 `json:"east"`
	West  float64 `json:"west"`
}
type LatLng []float64

func rectToBounds(rect s2.Rect) *Bound {
	return &Bound{
		North: rect.Hi().Lat.Degrees(),
		East:  rect.Hi().Lng.Degrees(),
		South: rect.Lo().Lat.Degrees(),
		West:  rect.Lo().Lng.Degrees(),
	}
}

func (b *Bound) String() string {
	return fmt.Sprintf("{\n\tnorth: %+v,\n\tsouth:%v,\n\teast:%v,\n\twest:%v\n}\n", b.North, b.South, b.East, b.West)
}

func getCells(loc *Locale) s2.CellUnion {
	origin := loc.Boundaries[0]
	rect := s2.RectFromLatLng(s2.LatLngFromDegrees(origin[0], origin[1]))
	for _, p := range loc.Boundaries[1:] {
		pp := s2.LatLngFromDegrees(p[0], p[1])
		rect = rect.AddPoint(pp)
	}

	rc := &s2.RegionCoverer{MinLevel: 15, MaxLevel: 15, MaxCells: 35}
	r := s2.Region(rect.CapBound())

	var cells s2.CellUnion
	for _, c := range rc.Covering(r) {
		cell := s2.CellFromCellID(c)
		if rect.IntersectsCell(cell) {
			cells = append(cells, c)
		}
	}

	return cells
}
