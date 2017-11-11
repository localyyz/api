package search

import (
	"fmt"
	"net/http"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/gedex/inflector"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	db "upper.io/db.v3"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/s", OmniSearch)
	r.Post("/category", SearchCategory)
	r.Post("/city", SearchCity)

	return r
}

func SearchCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cursor := api.NewPage(r)

	searchQuery := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	searchQuery = inflector.Singularize(searchQuery)

	// search the category tags that match the query term
	query := data.DB.ProductTag.Find(db.Cond{
		"value ~*": fmt.Sprint("\\m(", searchQuery, ")"),
		"type":     data.ProductTagTypeCategory,
	}).OrderBy("-created_at")
	query = cursor.UpdateQueryUpper(query)

	var tags []*data.ProductTag
	if err := query.All(&tags); err != nil {
		render.Respond(w, r, err)
		return
	}
	productIDs := make([]int64, len(tags))
	for i, t := range tags {
		productIDs[i] = t.ProductID
	}

	// find the products
	products, err := data.DB.Product.FindAll(db.Cond{"id": productIDs})
	if err != nil {
		render.Respond(w, r, err)
		return
	}
	render.RenderList(w, r, presenter.NewSearchProductList(ctx, products))

}

func SearchCity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cursor := api.NewPage(r)

	searchQuery := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	searchQuery = inflector.Singularize(searchQuery)

	// find the locale with search value
	locale, err := data.DB.Locale.FindOne(db.Cond{"shorthand": searchQuery})
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	// find the places with this relationship
	query := data.DB.
		Select(db.Raw("pl.*")).
		From("places pl").
		LeftJoin("place_locales pll").
		On("pl.id = pll.place_id").
		Where("pll.locale_id = ?", locale.ID).
		OrderBy("-pl.created_at")
	query = cursor.UpdateQueryBuilder(query)
	var places []*data.Place
	if err := query.All(&places); err != nil {
		render.Respond(w, r, err)
		return
	}
	render.RenderList(w, r, presenter.NewPlaceList(ctx, places))
}

// OmniSearch catch all search endpoint and returns categorized
// json search results
func OmniSearch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	searchQuery := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	searchQuery = inflector.Singularize(searchQuery)

	// find products by title
	query := data.DB.Product.Find(db.Cond{
		"title ~*": fmt.Sprint("\\m(", searchQuery, ")"),
	}).OrderBy("-created_at")
	cursor := api.NewPage(r)
	query = cursor.UpdateQueryUpper(query)

	var products []*data.Product
	if err := query.All(&products); err != nil {
		render.Respond(w, r, err)
		return
	}
	render.RenderList(w, r, presenter.NewSearchProductList(ctx, products))
}
