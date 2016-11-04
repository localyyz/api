package place

import (
	"net/http"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
)

func Search(w http.ResponseWriter, r *http.Request) {
}

// AutoComplete returns list of place names based on given query input
func AutoComplete(w http.ResponseWriter, r *http.Request) {
	queryString := strings.TrimSpace(r.URL.Query().Get("q"))
	places, err := data.DB.Place.FindLikeName(queryString)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}
	ws.Respond(w, http.StatusOK, places)
}
