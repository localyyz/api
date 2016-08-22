package place

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	db "upper.io/db.v2"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
	"bitbucket.org/moodie-app/moodie-api/web/utils"
	"github.com/pressly/chi"
)

func PlaceCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		placeID, err := strconv.ParseInt(chi.URLParam(r, "placeID"), 10, 64)
		if err != nil {
			ws.Respond(w, http.StatusBadRequest, utils.ErrBadID)
			return
		}

		place, err := data.DB.Place.FindByID(placeID)
		if err != nil {
			ws.Respond(w, http.StatusInternalServerError, err)
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, "place", place)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func GetPlace(w http.ResponseWriter, r *http.Request) {
	place := r.Context().Value("place").(*data.Place)

	locale, err := data.DB.Locale.FindByID(place.LocaleID)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	resp := &data.PlaceWithLocale{
		Place:  place,
		Locale: locale,
	}
	ws.Respond(w, http.StatusOK, resp)
}

// AutoCompletePlaces returns matched places via a query string
func AutoCompletePlaces(w http.ResponseWriter, r *http.Request) {
	queryString := strings.TrimSpace(r.URL.Query().Get("q"))
	places, err := data.DB.Place.Autocomplete(queryString)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}
	ws.Respond(w, http.StatusOK, places)
}

// NearbyPlaces returns places and promos based on user's last recorded geolocation
func NearbyPlaces(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)

	q := data.DB.Place.
		Find(db.Cond{"locale_id": user.Etc.LocaleID}).
		Select(
			db.Raw("*"),
			db.Raw(fmt.Sprintf("ST_Distance(geo, st_geographyfromtext('%v'::text)) distance", user.Geo)),
		).
		OrderBy("distance")

	var places []*data.Place
	if err := q.All(&places); err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}
	var resp struct {
		Places []*data.Place `json:"places"`
		Promos []*data.Promo `json:"promos"`
	}

	if len(places) == 0 {
		resp.Places = []*data.Place{}
		resp.Promos = []*data.Promo{}
		ws.Respond(w, http.StatusOK, resp)
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
	resp.Places = places
	resp.Promos = promos

	ws.Respond(w, http.StatusOK, resp)
}
