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
