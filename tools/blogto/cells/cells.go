package cells

import (
	"fmt"

	"bitbucket.org/moodie-app/moodie-api/data"
	db "upper.io/db.v2"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

func Analyze() {
	//dbCells, _ := data.DB.Cell.FindAll(db.Cond{"locale_id": 9})

	fmt.Println("go")
	p := s2.LatLngFromDegrees(43.650765, -79.387504)
	origin := s2.CellIDFromLatLng(p).Parent(16)

	cond := db.Cond{
		"cell_id >=": int(origin.RangeMin()),
		"cell_id <=": int(origin.RangeMax()),
	}
	cells, _ := data.DB.Cell.FindAll(cond)

	// min?
	min := s1.InfAngle()
	var localeID int64
	for _, c := range cells {

		cell := s2.CellID(c.CellID)
		d := p.Distance(cell.LatLng())
		fmt.Println(int(cell), c.LocaleID, p.Distance(cell.LatLng()))
		if d < min {
			min = d
			localeID = c.LocaleID
		}

	}
	fmt.Println(localeID)
}
