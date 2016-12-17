package locale

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	db "upper.io/db.v2"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/pkg/errors"
	"github.com/pressly/chi"
)

func LocaleCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		localeID, err := strconv.ParseInt(chi.URLParam(r, "localeID"), 10, 64)
		if err != nil {
			ws.Respond(w, http.StatusBadRequest, api.ErrBadID)
			return
		}

		locale, err := data.DB.Locale.FindByID(localeID)
		if err != nil {
			ws.Respond(w, http.StatusInternalServerError, err)
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, "locale", locale)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func ListLocale(w http.ResponseWriter, r *http.Request) {
	var locales []*data.Locale

	// TODO: for now, we're hackers
	err := data.DB.Locale.Find(
		data.EnabledLocales,
	).OrderBy("shorthand").All(&locales)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, errors.Wrap(err, "list locale"))
		return
	}
	ws.Respond(w, http.StatusOK, locales)
}

func ListPlaces(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	locale := ctx.Value("locale").(*data.Locale)
	user := ctx.Value("session.user").(*data.User)

	var places []*data.Place
	err := data.DB.Place.
		Find(db.Cond{"locale_id": locale.ID}).
		Select(
			db.Raw("*"),
			db.Raw(fmt.Sprintf("ST_Distance(geo, st_geographyfromtext('%v'::text)) distance", user.Geo)),
		).
		OrderBy("distance").
		Limit(20).
		All(&places)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, errors.Wrap(err, "list place"))
		return
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
