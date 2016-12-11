package place

import (
	"net/http"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
)

// AutoComplete returns list of place names based on given query input
func AutoComplete(w http.ResponseWriter, r *http.Request) {
	queryString := strings.TrimSpace(r.URL.Query().Get("q"))
	places, err := data.DB.Place.FindLikeName(queryString)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}
	presented := make([]*presenter.Place, len(places))
	for i, pl := range places {
		presented[i] = presenter.NewPlace(r.Context(), pl).WithGeo().WithLocale()
	}
	ws.Respond(w, http.StatusOK, presented)
}
