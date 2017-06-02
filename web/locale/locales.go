package locale

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/pressly/chi"
	"github.com/pressly/chi/render"
)

func LocaleCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		localeID, err := strconv.ParseInt(chi.URLParam(r, "localeID"), 10, 64)
		if err != nil {
			render.Render(w, r, api.ErrBadID)
			return
		}

		locale, err := data.DB.Locale.FindByID(localeID)
		if err != nil {
			render.Render(w, r, api.WrapErr(err))
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, "locale", locale)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func ListLocale(w http.ResponseWriter, r *http.Request) {
	// TODO: for now, we're hackers
	var locales []*data.Locale
	err := data.DB.Locale.Find(
		data.EnabledLocales,
	).OrderBy("shorthand").All(&locales)
	if err != nil {
		render.Render(w, r, api.WrapErr(err))
		return
	}

	presented := presenter.NewLocaleList(r.Context(), locales)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Render(w, r, api.WrapErr(err))
	}
}

func ListPlaces(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	locale := ctx.Value("locale").(*data.Locale)
	user := ctx.Value("session.user").(*data.User)

	cursor := ws.NewPage(r)

	var places []*data.Place
	query := data.DB.Place.
		Find(db.Cond{"locale_id": locale.ID}).
		Select(
			db.Raw("*"),
			db.Raw(fmt.Sprintf("ST_Distance(geo, st_geographyfromtext('%v'::text)) distance", user.Geo)),
		).
		OrderBy("distance")

	query = cursor.UpdateQueryUpper(query)
	if err := query.All(&places); err != nil {
		render.Render(w, r, api.WrapErr(err))
		return
	}

	presented := presenter.NewPlaceList(ctx, places)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Render(w, r, nil)
	}
}
