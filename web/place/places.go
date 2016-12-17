package place

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"upper.io/db.v2"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/pkg/errors"
	"github.com/pressly/chi"
)

func PlaceCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		placeID, err := strconv.ParseInt(chi.URLParam(r, "placeID"), 10, 64)
		if err != nil {
			ws.Respond(w, http.StatusBadRequest, api.ErrBadID)
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

// getTrending returns most popular places ordered by aggregated score
func getTrending(user *data.User) ([]*data.Place, error) {
	var places []*data.Place
	q := data.DB.
		Select(
			db.Raw("pl.*"),
			db.Raw(fmt.Sprintf("ST_Distance(pl.geo, st_geographyfromtext('%v'::text)) distance", user.Geo)),
		).
		From("places pl").
		LeftJoin("claims cl").
		On("pl.id = cl.place_id").
		GroupBy("pl.id").
		OrderBy(db.Raw("count(cl) DESC NULLS LAST")).
		Limit(10)
	if err := q.All(&places); err != nil {
		return nil, errors.Wrap(err, "trending places")
	}

	return places, nil
}

func ListPlaces(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUser := ctx.Value("session.user").(*data.User)

	var places []*data.Place
	// if not admin, return
	if !currentUser.IsAdmin {
		ws.Respond(w, http.StatusOK, places)
		return
	}

	var err error
	places, err = data.DB.Place.FindAll(db.Cond{})
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}
	ws.Respond(w, http.StatusOK, places)
}

func GetPlace(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)
	ws.Respond(w, http.StatusOK, (presenter.NewPlace(ctx, place)).WithGeo().WithLocale())
}

// Nearby returns places and promos based on user's last recorded geolocation
func Nearby(w http.ResponseWriter, r *http.Request) {
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
		//Limit(20).
		All(&places)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	if len(places) == 0 {
		places, err = getTrending(user)
		if err != nil {
			ws.Respond(w, http.StatusInternalServerError, err)
			return
		}
	}

	var presented []*presenter.Place
	for _, pl := range places {
		// TODO: +1 here
		p := presenter.NewPlace(ctx, pl).WithPromo()
		if p.Promo.Promo == nil {
			continue
		}
		presented = append(presented, p.WithLocale().WithGeo())
	}

	ws.Respond(w, http.StatusOK, presented)
}
