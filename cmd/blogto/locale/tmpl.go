package locale

func ShowLocale() {
	//loc := cache[*localeID] // liberty village
	//origin := loc.Boundaries[0]
	//rect := s2.RectFromLatLng(s2.LatLngFromDegrees(origin[0], origin[1]))

	//maps, err := template.New("maps").Parse(tmpl)
	//if err != nil {
	//io.WriteString(w, err.Error())
	//return
	//}

	//var cellBounds []*Bound
	//for _, c := range getCells(&loc) {
	//cell := s2.CellFromCellID(c)
	//cellBounds = append(cellBounds, rectToBounds(cell.RectBound()))
	//}

	//z := struct {
	//CenterLat float64
	//CenterLng float64
	//Rect      *Bound
	//Cells     []*Bound
	//}{
	//rect.Center().Lat.Degrees(),
	//rect.Center().Lng.Degrees(),
	//rectToBounds(rect),
	//cellBounds,
	//}

	//b := &bytes.Buffer{}
	//if err := maps.Execute(b, z); err != nil {
	//io.WriteString(w, err.Error())
	//return
	//}

	//w.Header().Set("Content-Type", "text/html")
	//w.Write(b.Bytes())
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
</html>`
