package place

import (
	"context"
	"net/http"
	"strings"

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

	ws.Respond(w, http.StatusOK, places)
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
