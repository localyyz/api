package locale

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"

	db "upper.io/db.v2"

	"bitbucket.org/moodie-app/moodie-api/data"

	"github.com/golang/geo/s2"
	"github.com/gosimple/slug"
	"github.com/goware/lg"
	"github.com/pkg/errors"
)

type Locale struct {
	ID         int64 `json:"id"`
	Boundaries struct {
		Type        string     `json:"type"`
		Coordinates [][]LatLng `json:"coordinates"`
	} `json:"boundaries"`
	ImageURL  string `json:"image_url"`
	Blurb     string `json:"blurb"`
	Name      string `json:"name"`
	Shorthand string
}
type Locales []*Locale

var (
	cache     Locales
	LocaleMap map[string]*Locale
)

func (slice Locales) Len() int {
	return len(slice)
}

func (slice Locales) Less(i, j int) bool {
	return slice[i].ID < slice[j].ID
}

func (slice Locales) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func List() {
	localeList := ""
	for _, loc := range cache {
		localeList += fmt.Sprintf(" %d", loc.ID)
		localeList += fmt.Sprintf(" %s", loc.Name)
		localeList += fmt.Sprintf(" (%s) \n", loc.Shorthand)
	}
	fmt.Println(localeList)
}

func (loc *Locale) GetCoords() ([]LatLng, error) {
	if len(loc.Boundaries.Coordinates) == 0 ||
		len(loc.Boundaries.Coordinates[0]) == 0 {
		return nil, errors.New("no coords")
	}

	// weird 1 element list
	return loc.Boundaries.Coordinates[0], nil
}

func (loc *Locale) GetBoundaryCells() s2.CellUnion {
	coords, err := loc.GetCoords()
	if err != nil {
		return nil
	}

	cells, err := GetS2Cover(coords)
	if err != nil {
		panic(err)
	}

	//var cells s2.CellUnion
	//rc := NewFlatCoverer(15)

	//for _, p := range coords {
	//rect := s2.RectFromLatLng(s2.LatLngFromDegrees(p[1], p[0]))
	//rect = rect.AddPoint(s2.LatLngFromDegrees(p[1], p[0]))
	//for _, c := range rc.Covering(r) {
	//neighbours := c.EdgeNeighbors()
	//cells = append(cells, neighbours[0], neighbours[1], neighbours[2], neighbours[3])
	//}
	//r := s2.Region(rect)
	//cells = append(cells, rc.Covering(r)...)
	//}

	// info:
	// 1. http://blog.christianperone.com/2015/08/googles-s2-geometry-on-the-sphere-cells-and-hilbert-curve/
	// 2. https://medium.com/@buckhx/unwinding-uber-s-most-efficient-service-406413c5871d#.msvnff7sv
	// 3. https://github.com/buckhx/gofence
	// 4. https://github.com/blackmad/s2map

	// note: apparently. there's an api for this all.

	//origin := coords[0]
	//rect := s2.RectFromLatLng(s2.LatLngFromDegrees(origin[1], origin[0]))
	//for _, p := range coords[1:] {
	//pp := s2.LatLngFromDegrees(p[1], p[0])
	//rect = rect.AddPoint(pp)
	//break
	//}

	//r := s2.Region(rect.CapBound())

	//var cells s2.CellUnion
	//for _, c := range rc.Covering(r) {
	//cell := s2.CellFromCellID(c)
	//if rect.IntersectsCell(cell) {
	//cells = append(cells, c)
	//}
	//}

	// testing region coverer
	//return rc.Covering(r)

	return cells
}

func LoadLocale() {
	for sh, loc := range LocaleMap {
		// check if already exists
		dbLocale, err := data.DB.Locale.FindOne(db.Cond{"shorthand": sh})
		if err != nil {
			if err != db.ErrNoMoreRows {
				log.Fatalf("loading error: %+v", err)
				continue
			}
			dbLocale = &data.Locale{
				Name:        loc.Name,
				Shorthand:   sh,
				Description: loc.Blurb,
			}
			if err := data.DB.Locale.Save(dbLocale); err != nil {
				log.Fatal(err)
			}
		}

		lg.Infof("Loading cells for %s", sh)
		for _, c := range loc.GetBoundaryCells() {
			cell := &data.Cell{LocaleID: dbLocale.ID, CellID: int64(c)}
			if err := data.DB.Cell.Save(cell); err != nil {
				lg.Warn(errors.Wrapf(err, "failed inserting cell(%v)", int64(c)))
				break
			}
		}
	}
}

func init() {
	// TODO: make into config
	f, err := os.Open("./data/locale.json")
	if err != nil {
		log.Fatal(err)
	}

	buf, _ := ioutil.ReadAll(f)
	if err := json.Unmarshal(buf, &cache); err != nil {
		log.Fatal(err)
	}
	sort.Sort(cache)

	// write to map for easy lookup
	LocaleMap = make(map[string]*Locale, len(cache))
	for _, l := range cache {
		l.Shorthand = slug.Make(l.Name)
		LocaleMap[l.Shorthand] = l
	}
}
