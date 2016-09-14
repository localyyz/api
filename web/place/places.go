package place

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"upper.io/db.v2"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
	"bitbucket.org/moodie-app/moodie-api/web/utils"
	"github.com/pkg/errors"
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
// TODO: present distance
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

	var places []*data.Place
	err := data.DB.Place.
		Find(db.Cond{"locale_id": user.Etc.LocaleID}).
		Select(
			db.Raw("*"),
			db.Raw(fmt.Sprintf("ST_Distance(geo, st_geographyfromtext('%v'::text)) distance", user.Geo)),
		).
		OrderBy("distance").
		All(&places)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	presented, err := presenter.PlacesWithPromos(ctx, places...)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	ws.Respond(w, http.StatusOK, presented)
}

// ListTrendingNearby returns most popular places
//  ordered by aggregated score
func ListTrending(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	q := data.DB.Select(db.Raw("pl.*")).
		From("places pl").
		Join("posts p").
		On("pl.id = p.place_id").
		GroupBy("pl.id").
		OrderBy(db.Raw("sum(p.score) DESC")).
		Limit(15)
	var places []*data.Place
	if err := q.All(&places); err != nil {
		ws.Respond(w, http.StatusInternalServerError, errors.Wrap(err, "trending places"))
		return
	}

	presented, err := presenter.PlacesWithPosts(ctx, places...)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}
	ws.Respond(w, http.StatusOK, presented)
}

// PeekPromo let's the user take a peek at the promotion
//  that might be too far away. For a price of course.
func PeekPromo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)
	place := ctx.Value("place").(*data.Place)

	promo, err := data.DB.Promo.FindOne(db.Cond{"place_id": place.ID})
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	// don't need to peek
	if place.Distance < data.PromoDistanceLimit {
		ws.Respond(w, http.StatusOK, promo)
		return
	}

	peek := &data.PromoPeek{UserID: user.ID, PromoID: promo.ID}
	err = data.DB.PromoPeek.Save(peek)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	err = data.DB.UserPoint.Save(
		&data.UserPoint{
			UserID:  user.ID,
			PlaceID: place.ID,
			PromoID: promo.ID,
			PeekID:  peek.ID,
			Reward:  -promo.Reward / 10, // hardcode to 10th of the reward
		},
	)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	ws.Respond(w, http.StatusOK, promo)
}
