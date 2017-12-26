package search

import (
	"fmt"
	"net/http"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"bitbucket.org/moodie-app/moodie-api/web/locale"
	"github.com/gedex/inflector"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	db "upper.io/db.v3"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", OmniSearch)
	r.With(locale.LocaleShorthandCtx).Post("/city/{locale}", SearchCity)

	return r
}

func SearchCity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// find the locale with search value
	locale := ctx.Value("locale").(*data.Locale)

	// find the places with this relationship
	query := data.DB.Place.Find(
		db.Cond{
			"locale_id": locale.ID,
			"status":    data.PlaceStatusActive,
		},
	).OrderBy("-id")
	cursor := ctx.Value("cursor").(*api.Page)
	query = cursor.UpdateQueryUpper(query)
	var places []*data.Place
	if err := query.All(&places); err != nil {
		render.Respond(w, r, err)
		return
	}
	cursor.Update(places)
	render.RenderList(w, r, presenter.NewPlaceList(ctx, places))
}

type omniSearchRequest struct {
	Query string `json:"query,required"`

	singularQuery string
	searchQuery   string
}

func (o *omniSearchRequest) Bind(r *http.Request) error {
	o.searchQuery = strings.ToLower(strings.TrimSpace(o.Query))
	o.singularQuery = inflector.Singularize(o.searchQuery)

	return nil
}

// OmniSearch catch all search endpoint and returns categorized
// json search results
func OmniSearch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var p omniSearchRequest
	if err := render.Bind(r, &p); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	// find products by title
	query := data.DB.Product.Find(
		db.Or(
			db.Cond{"title ~*": fmt.Sprint("\\m(", p.singularQuery, ")")},
			db.Cond{"title ~*": fmt.Sprint("\\m(", p.searchQuery, ")")},
			db.Raw("etc->>'brand' ~* ?", fmt.Sprint("\\m(", p.singularQuery, ")")),
			db.Raw("etc->>'brand' ~* ?", fmt.Sprint("\\m(", p.searchQuery, ")")),
		),
	).OrderBy("-created_at")
	cursor := ctx.Value("cursor").(*api.Page)
	query = cursor.UpdateQueryUpper(query)

	var products []*data.Product
	if err := query.All(&products); err != nil {
		render.Respond(w, r, err)
		return
	}
	cursor.Update(products)
	render.RenderList(w, r, presenter.NewSearchProductList(ctx, products))
}
