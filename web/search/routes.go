package search

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/gedex/inflector"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	set "gopkg.in/fatih/set.v0"
	db "upper.io/db.v3"
)

func Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", OmniSearch)

	return r
}

type omniSearchRequest struct {
	Query string `json:"query,required"`

	queryParts []string
}

var alphanumRx = regexp.MustCompile("[^a-zA-Z0-9-]+")

func (o *omniSearchRequest) Bind(r *http.Request) error {
	if len(o.Query) == 0 {
		return errors.New("invalid empty search query")
	}

	qSet := set.New()
	for _, t := range alphanumRx.Split(strings.ToLower(o.Query), -1) {
		tt := inflector.Singularize(t)
		for {
			if tt == t {
				break
			}
			t = tt
			tt = inflector.Singularize(t)
		}
		if t == "" {
			continue
		}
		qSet.Add(t)
	}
	if qSet.Size() == 0 {
		return errors.New("invalid search query")
	}
	o.queryParts = set.StringSlice(qSet)

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

	// join query parts back to one string
	qraw := db.Raw(strings.Join(p.queryParts, " "))
	whereClause := db.Raw(`tsv @@ plainto_tsquery($$?$$) and deleted_at is null`, qraw)

	// find products by title
	query := data.DB.Select(
		"*",
		db.Raw("ts_rank(tsv, plainto_tsquery($$?$$)) + weight as rank", qraw)).
		From("products").
		Where(whereClause).
		OrderBy("rank DESC", "id DESC")

	cursor := ctx.Value("cursor").(*api.Page)
	{
		count, _ := data.DB.Product.Find(whereClause).Count()
		cursor.ItemTotal = int(count)
	}

	paginate := cursor.UpdateQueryBuilder(query)
	var products []*data.Product
	if err := paginate.All(&products); err != nil {
		render.Respond(w, r, err)
		return
	}
	cursor.Update(products)
	render.RenderList(w, r, presenter.NewSearchProductList(ctx, products))
}
