package locale

import (
	"bytes"
	"html/template"
	"io"
	"net/http"

	"github.com/golang/geo/s2"
	"github.com/goware/lg"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
	"upper.io/db.v3"
)

func LocaleHandler(w http.ResponseWriter, r *http.Request) {
	maps, err := template.New("maps").Parse(tmpl)
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}

	sh := r.URL.Query().Get("sh")
	lg.Warnf("generating cells for %s", sh)
	locale, err := data.DB.Locale.FindOne(db.Cond{"shorthand": sh})
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	cells, err := data.DB.Cell.FindAll(db.Cond{"locale_id": locale.ID})
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	var cellBounds []*Bound
	for _, c := range cells {
		cell := s2.CellFromCellID(s2.CellID(c.CellID))
		cellBounds = append(cellBounds, rectToBounds(cell.RectBound()))
	}

	// center the map on the first point
	z := struct {
		CenterLat float64
		CenterLng float64
		Cells     []*Bound
	}{
		cellBounds[0].North,
		cellBounds[0].East,
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
			}
			</script>
	</body>
</html>`
