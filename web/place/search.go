package place

import (
	"context"
	"net/http"
	"strings"

	"upper.io/db"

	"googlemaps.github.io/maps"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
)

func PlaceTypeCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		placeType, err := maps.ParsePlaceType(r.URL.Query().Get("t"))
		if err != nil {
			ws.Respond(w, http.StatusBadRequest, err)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, "place.type", placeType)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func NearbyPlaces(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)

	places, err := data.GetNearby(ctx, &user.Geo)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	// return any active promotions
	placeIDs := make([]int64, len(places))
	for i, p := range places {
		placeIDs[i] = p.ID
	}

	// query promos
	promos, err := data.DB.Promo.FindAll(db.Cond{"place_id IN": placeIDs})
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	// places with promos
	resp := struct {
		Places []*data.Place `json:"places"`
		Promos []*data.Promo `json:"promos"`
	}{
		places,
		promos,
	}

	ws.Respond(w, http.StatusOK, resp)
}

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
