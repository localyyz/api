package locale

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
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
			render.Respond(w, r, err)
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, "locale", locale)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func ListCities(w http.ResponseWriter, r *http.Request) {
	locales, err := data.DB.Locale.FindAll(db.Cond{"type": data.LocaleTypeCity})
	if err != nil {
		render.Respond(w, r, err)
		return
	}
	render.Respond(w, r, locales)
}

func ListPlaces(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	locale := ctx.Value("locale").(*data.Locale)
	user := ctx.Value("session.user").(*data.User)

	cursor := api.NewPage(r)

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
		render.Respond(w, r, err)
		return
	}

	presented := presenter.NewPlaceList(ctx, places)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}
}
