package place

import (
	"net/http"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
)

// SearchPlaces takes a place name and returns all the places that match from
// google maps api
func SearchPlaces(w http.ResponseWriter, r *http.Request) {
	var places []*data.Place

	placeName := strings.TrimSpace(r.URL.Query().Get("q"))
	if placeName == "" {
		ws.Respond(w, http.StatusOK, places)
	}

	// look up the place

	ws.Respond(w, http.StatusOK, places)
}
