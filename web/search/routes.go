package search

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/gedex/inflector"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
	set "gopkg.in/fatih/set.v0"
	db "upper.io/db.v3"
	"upper.io/db.v3/lib/sqlbuilder"
)

func Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", OmniSearch)
	r.Post("/related", RelatedTags)

	return r
}

type omniSearchRequest struct {
	Query string `json:"query,required"`
	// Tags to filter out
	FilterTags []string `json:"filterTags,omitempty"`
	queryParts []string

	rawParts []string
}

func (o *omniSearchRequest) Bind(r *http.Request) error {
	if len(o.Query) == 0 {
		return errors.New("invalid empty search query")
	}

	qSet := set.New()
	for _, t := range strings.Split(strings.ToLower(o.Query), " ") {
		o.rawParts = append(o.rawParts, t)

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
		// set up for partial prefix matching
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
	qraw := db.Raw(strings.Join(p.queryParts, ":* &"))
	qrawNoSpace := strings.Join(p.queryParts, "")

	tquery := db.Raw(`tsv @@ (
		to_tsquery($$?$$) ||
		to_tsquery('simple', $$?:*$$) ||
		to_tsquery($$?:*$$) ||
		to_tsquery('simple', $$?$$)
	)`, qraw, qraw, qraw, db.Raw(qrawNoSpace))

	cursor := ctx.Value("cursor").(*api.Page)
	{
		// TODO: filter products not in stock
		// NOTE: the trade off between a real count (filtering products out of
		// stock) and a "good enough" count in terms of query time is 10x. Not
		// worth it.
		cond := db.Cond{
			"p.deleted_at IS": nil,
			"p.status":        data.ProductStatusApproved,
			"pl.status":       data.PlaceStatusActive,
		}
		if gender, ok := ctx.Value("session.gender").(data.UserGender); ok {
			cond["p.gender"] = gender
		}

		row, err := data.DB.Select(
			db.Raw("count(1) AS _t")).
			From("products p").
			LeftJoin("places pl").On("pl.id = p.place_id").
			Where(db.And(cond, tquery)).
			QueryRow()
		var count uint64
		if err == nil {
			row.Scan(&count)
		}
		cursor.ItemTotal = int(count)
	}

	var query sqlbuilder.Selector
	if cursor.ItemTotal > 0 {
		// NOTE on magic numbers
		//
		// ranking type 32 => normalizes the rank with `x / (1+x)` where x is the original rank
		// modifier 4 => is the top 70th (another magical number) of our merchant
		//      weights greater than 0
		cond := db.Cond{
			"p.deleted_at IS": nil,
			"pl.status":       data.PlaceStatusActive,
			"p.status":        data.ProductStatusApproved,
		}
		if gender, ok := ctx.Value("session.gender").(data.UserGender); ok {
			cond["p.gender"] = gender
		}
		query = data.DB.Select(
			db.Raw("distinct p.id"),
			db.Raw(data.ProductQueryWeight, qraw)).
			From("products p").
			LeftJoin("places pl").On("pl.id = p.place_id").
			Where(db.And(cond, tquery)).
			OrderBy("_rank DESC")
	} else {
		// find best matched spellings for each word in the query
		// NOTE: make sure to fuzzy search with raw and unparsed query
		//
		// for example: `addidas` shouldn't be inflected to `addida`
		fuzzyWords, _ := data.DB.SearchWord.FindSimilar(p.rawParts...)

		// if we didn't find any fuzzyWords, return
		if len(fuzzyWords) == 0 {
			render.Respond(w, r, []struct{}{})
			return
		}
		// for search queries, we have the default weighting vector of {0.1, 0.2, 0.4, 1.0}
		// which we can use to specify the importance of each terms in a query.

		// for example if user searches for "yeezy shoes"
		// the two distinct terms "yeezy" and "shoes" shouldn't carry the same
		// weight in the query, because "yeezy" is a much better representation
		// of what the user is actually searching for.

		// handle this by normalizing the ndoc occurance of search term and
		// mapping them to the ranking of {D, C, B, A}

		// for now, it's a nieve mapping
		// lowest ndoc occuring search term is the A value, rest are B

		// sort fuzzyWords by number of occurences, smallest to largest
		fuzzySorter := data.WordFrequencySorter(fuzzyWords)
		sort.Sort(fuzzySorter)

		// for each fuzzyWords, search again in products for the "corrected" query
		var andTerm, orTerm string
		for i, w := range fuzzySorter {
			if i == 0 {
				orTerm = fmt.Sprintf("%s", w.Word)
				andTerm = fmt.Sprintf("%s:B", w.Word)
				continue
			}
			andTerm += fmt.Sprintf(" & %s:B", w.Word)
		}

		cond := db.Cond{
			"p.deleted_at IS": nil,
			"p.status":        data.ProductStatusApproved,
			"pl.status":       data.PlaceStatusActive,
		}
		if gender, ok := ctx.Value("session.gender").(data.UserGender); ok {
			cond["p.gender"] = gender
		}

		term := fmt.Sprintf("%s | %s", orTerm, andTerm)
		query = data.DB.Select(
			db.Raw("distinct p.id"),
			db.Raw(data.ProductFuzzyWeight, term)).
			From("products p").
			LeftJoin("places pl").On("pl.id = p.place_id").
			Where(
				db.And(db.Raw(`tsv @@ to_tsquery(?)`, term), cond),
			).
			OrderBy("_rank DESC")
	}
	paginate := cursor.UpdateQueryBuilder(query)

	rows, err := paginate.QueryContext(ctx)
	if err != nil {
		render.Respond(w, r, err)
		return
	}
	defer rows.Close()

	var productIDs []int64
	for rows.Next() {
		var pId int64
		var rank interface{}
		if err := rows.Scan(&pId, &rank); err != nil {
			lg.Warnf("error scanning query: %+v", err)
			break
		}
		productIDs = append(productIDs, pId)
	}
	if err := rows.Err(); err != nil {
		render.Respond(w, r, err)
		return
	}

	result := data.DB.Product.Find(
		db.Cond{"id": productIDs},
	).OrderBy(
		data.MaintainOrder("id", productIDs),
	)
	var products []*data.Product
	if err := result.All(&products); err != nil {
		render.Respond(w, r, err)
		return
	}
	cursor.Update(products)

	render.RenderList(w, r, presenter.NewSearchProductList(ctx, products))
}
