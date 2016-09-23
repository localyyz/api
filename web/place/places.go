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
		Limit(20).
		All(&places)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	presented := make([]*presenter.Place, len(places))
	for i, pl := range places {
		presented[i] = presenter.NewPlace(ctx, pl).WithLocale().WithGeo().WithPromo()
	}

	ws.Respond(w, http.StatusOK, presented)
}

// Trending returns most popular places ordered by aggregated score
func Trending(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)

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
		Where(db.Cond{"pl.locale_id !=": user.Etc.LocaleID}).
		Limit(10)
	if err := q.All(&places); err != nil {
		ws.Respond(w, http.StatusInternalServerError, errors.Wrap(err, "trending places"))
		return
	}

	presented := make([]*presenter.Place, len(places))
	for i, pl := range places {
		presented[i] = presenter.NewPlace(ctx, pl).WithLocale().WithGeo().WithPromo()
	}
	ws.Respond(w, http.StatusOK, presented)
}
