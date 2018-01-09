package locale

import (
	"context"
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

func LocaleShorthandCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		shorthand := chi.URLParam(r, "locale")
		if len(shorthand) == 0 {
			render.Render(w, r, api.ErrBadID)
			return
		}

		locale, err := data.DB.Locale.FindOne(db.Cond{"shorthand": shorthand})
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
	var locales []*data.Locale
	err := data.DB.
		Select("l.*").
		From("locales l").
		LeftJoin("places p").
		On("p.locale_id = l.id").
		Where(db.Cond{"type": data.LocaleTypeCity, "status": data.PlaceStatusActive}).
		GroupBy("l.id").
		OrderBy(db.Raw("count(p) desc")).
		All(&locales)
	if err != nil {
		render.Respond(w, r, err)
		return
	}
	render.Respond(w, r, locales)
}

func ListPlaces(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	locale := ctx.Value("locale").(*data.Locale)

	cursor := api.NewPage(r)

	var places []*data.Place
	query := data.DB.Place.Find(db.Cond{"locale_id": locale.ID})
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
