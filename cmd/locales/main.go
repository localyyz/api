package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"bitbucket.org/moodie-app/moodie-api/data"
	"upper.io/db.v2"

	"github.com/golang/geo/s2"
	"github.com/goware/lg"
	"github.com/pkg/errors"
)

type Locale struct {
	ID          int64    `json:"id"`
	Boundaries  []LatLng `json:"boundaries"`
	Description string   `json:"description"`
	Name        string   `json:"name"`
	Shorthand   string   `json:"shorthand"`
}
type Bound struct {
	North float64 `json:"north"`
	South float64 `json:"south"`
	East  float64 `json:"east"`
	West  float64 `json:"west"`
}

type LatLng []float64

var (
	flags      = flag.NewFlagSet("locale", flag.ExitOnError)
	doLoad     = flags.Bool("load", false, "Load a locale into db")
	listLocale = flags.Bool("list", false, "Print the locale list and their id")
	localeID   = flags.Int64("id", 0, "Locale id to load")
	locales    map[int64]Locale
)

const tmpl = `
<html>
	<head>
		<style type="text/css">
			html, body {
				height: 100%;
				margin: 0;
				padding: 0;
			}
			#map {
				height: 100%;
			}
		</style>
	</head>
	<body>
		<div id="map"></div>
		<script async defer src="https://maps.googleapis.com/maps/api/js?key=AIzaSyBR-m6BuuHv3MDTuNCACiEOL4WA3l7gbzA&callback=initMap"></script>
		<script type="text/javascript">
			// This example adds a red rectangle to a map.
			function initMap() {
				var map = new google.maps.Map(document.getElementById('map'), {
					zoom: 15,
					center: {lat: {{ .CenterLat }}, lng: {{ .CenterLng }}},
					mapTypeId: 'terrain'
				});
				{{ range $i, $e := .Cells }}
				new google.maps.Rectangle({
					strokeColor: '#FF0000',
					strokeOpacity: 0.8,
					strokeWeight: 2,
					fillColor: '#FF0000',
					fillOpacity: 0.35,
					map: map,
					bounds :{
							north: {{ $e.North }},
							south: {{ $e.South }},
							east: {{ $e.East }},
							west: {{ $e.West }}
					}
				});
				{{ end }}
				var queenwest = new google.maps.Rectangle({
					strokeColor: '#00FF00',
					strokeOpacity: 0.8,
					strokeWeight: 2,
					fillColor: '#00FF00',
					fillOpacity: 0.35,
					map: map,
					bounds: {
						north: {{ .Rect.North }},
						south: {{ .Rect.South }},
						east: {{ .Rect.East }},
						west: {{ .Rect.West }}
					}
				});
			}
			</script>
	</body>
</html>
`

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

func LocaleHandler(w http.ResponseWriter, req *http.Request) {
	loc := locales[*localeID] // liberty village
	origin := loc.Boundaries[0]
	rect := s2.RectFromLatLng(s2.LatLngFromDegrees(origin[0], origin[1]))

	maps, err := template.New("maps").Parse(tmpl)
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}

	var cellBounds []*Bound
	for _, c := range getCells(&loc) {
		cell := s2.CellFromCellID(c)
		cellBounds = append(cellBounds, rectToBounds(cell.RectBound()))
	}

	z := struct {
		CenterLat float64
		CenterLng float64
		Rect      *Bound
		Cells     []*Bound
	}{
		rect.Center().Lat.Degrees(),
		rect.Center().Lng.Degrees(),
		rectToBounds(rect),
		cellBounds,
	}

	b := &bytes.Buffer{}
	if err := maps.Execute(b, z); err != nil {
		io.WriteString(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(b.Bytes())
}

func loadData() {
	loc := locales[*localeID]

	var dbLocale *data.Locale
	err := data.DB.Locale.Find(db.Cond{"shorthand": loc.Shorthand}).One(&dbLocale)
	if err != nil {
		if err != db.ErrNoMoreRows {
			log.Fatal(err)
		}
		desc, _ := url.QueryUnescape(loc.Description)
		dbLocale = &data.Locale{
			Shorthand:   loc.Shorthand,
			Name:        loc.Name,
			Description: desc,
		}
		if err := data.DB.Locale.Save(dbLocale); err != nil {
			log.Fatal(err)
		}
	}
	lg.Warnf("found locale(%v): %s", dbLocale.ID, dbLocale.Name)

	for _, c := range getCells(&loc) {
		cell := &data.Cell{LocaleID: dbLocale.ID, CellID: int64(c)}
		if err := data.DB.Cell.Save(cell); err != nil {
			lg.Warn(errors.Wrapf(err, "failed inserting cell(%v)", int64(c)))
			break
		}
	}
}

func main() {
	flags.Parse(os.Args[1:])

	f, err := os.Open("cmd/locales/locale.json")
	if err != nil {
		log.Fatal(err)
	}

	b, _ := ioutil.ReadAll(f)
	var lc []Locale
	if err := json.Unmarshal(b, &lc); err != nil {
		log.Fatal(err)
	}

	locales = make(map[int64]Locale)
	for _, l := range lc {
		locales[l.ID] = l
	}

	conf := data.DBConf{
		Database: "moodie",
		Hosts:    []string{":5432"},
		Username: "moodie",
	}
	if err := data.NewDBSession(conf); err != nil {
		lg.Fatal(err)
	}

	if *doLoad {
		loadData()
	} else if *listLocale {
		localeList := ""
		for id, loc := range locales {
			localeList += fmt.Sprintf("\tID: %d", id)
			localeList += fmt.Sprintf("\t\tName: %s \n", loc.Name)
		}
		fmt.Println(localeList)
	} else {
		lg.Warn("Starting server on :1234")
		http.HandleFunc("/", LocaleHandler)
		http.ListenAndServe(":1234", nil)
	}
}
