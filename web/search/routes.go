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
)

func Routes() chi.Router {

	r := chi.NewRouter()
	r.Route("/", api.FilterRoutes(Search))
	r.Post("/similar", SimilarSearch)
	r.Post("/related", RelatedTags)

	return r
}

type keywordPartType uint32

const (
	_ keywordPartType = iota
	keywordPartTypeGender
	keywordPartTypeCategory
)

var (
	ErrNoSearchResult = errors.New("no search results")
)

type omniSearchRequest struct {
	Query string `json:"query,required"`
	// Tags to filter out
	FilterTags []string `json:"filterTags,omitempty"`
	queryParts []string

	// keyword parts used to generate queries from keywords
	gender   *data.ProductGender
	category *data.CategoryType

	rawParts []string
}

// filter tags for genders
func parseGender(t string) data.ProductGender {
	switch t {
	case "man", "male", "gentleman":
		return data.ProductGenderMale
	case "woman", "female", "lady":
		return data.ProductGenderFemale
	case "kid":
		return data.ProductGenderUnisex
	case "sexy":
		// maybe female.
		return data.ProductGenderFemale
	}
	return data.ProductGenderUnknown
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

		// check if this term is a "keyword"
		if v := parseGender(t); v != data.ProductGenderUnknown {
			o.gender = &v
			continue
		}

		//if v, ok := data.CategoryLookup[t]; ok {
		//o.category = &v
		//continue
		//}

		// set up for partial prefix matching
		qSet.Add(t)

	}
	if qSet.Size() == 0 {
		return errors.New("invalid search query")
	}
	o.queryParts = set.StringSlice(qSet)

	return nil
}

func SimilarSearch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cursor := ctx.Value("cursor").(*api.Page)
	filterSort := ctx.Value("filter.sort").(*api.FilterSort)

	var p omniSearchRequest
	if err := render.Bind(r, &p); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	// find best matched spellings for each word in the query
	// NOTE: make sure to fuzzy search with raw and unparsed query
	//
	// for example: `addidas` shouldn't be inflected to `addida`
	fuzzyWords, _ := data.DB.SearchWord.FindSimilar(p.rawParts...)
	// if we didn't find any fuzzyWords, return
	if len(fuzzyWords) == 0 {
		render.Respond(w, r, ErrNoSearchResult)
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
		"p.deleted_at": db.IsNull(),
		"p.status":     data.ProductStatusApproved,
		"p.score":      db.Gt(0),
		"pl.status":    data.PlaceStatusActive,
	}
	if gender, ok := ctx.Value("session.gender").(data.UserGender); ok {
		cond["p.gender"] = gender
	}

	term := fmt.Sprintf("%s | %s", orTerm, andTerm)
	query := data.DB.Select(
		db.Raw("distinct p.id"),
		db.Raw(data.ProductFuzzyWeight, term)).
		From("products p").
		LeftJoin("places pl").On("pl.id = p.place_id").
		Where(
			db.And(db.Raw(`tsv @@ to_tsquery(?)`, term), cond),
		).
		OrderBy("_rank DESC")
	query = filterSort.UpdateQueryBuilder(query)
	paginate := cursor.UpdateQueryBuilder(query)

	rows, err := paginate.QueryContext(ctx)
	if err != nil {
		lg.Warnf("search: failed with %v", err)
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

	render.RenderList(w, r, presenter.NewProductList(ctx, products))
}

// OmniSearch catch all search endpoint and returns categorized
// json search results
func Search(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cursor := ctx.Value("cursor").(*api.Page)
	filterSort := ctx.Value("filter.sort").(*api.FilterSort)

	var p omniSearchRequest
	if err := render.Bind(r, &p); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	// log the search query
	lg.SetEntryField(ctx, "query", p.Query)

	cond := db.And(db.Cond{
		"p.deleted_at": db.IsNull(),
		"p.status":     data.ProductStatusApproved,
		"p.score":      db.Gt(0),
	})
	if p.gender != nil {
		cond = cond.And(db.Cond{"p.gender": *p.gender})
	}
	if p.category != nil {
		cond = cond.And(db.Cond{"p.category->>'type'": p.category.String()})
	}
	// join query parts back to one string
	qraw := db.Raw(strings.Join(p.queryParts, ":* &"))
	qrawNoSpace := strings.Join(p.queryParts, "")
	cond = cond.And(db.Raw(`
		tsv @@ (
			to_tsquery($$?$$) ||
			to_tsquery('simple', $$?:*$$) ||
			to_tsquery($$?:*$$) ||
			to_tsquery('simple', $$?$$)
		)`, qraw, qraw, qraw, db.Raw(qrawNoSpace)))

	query := data.DB.Select("p.*").
		From("products p").
		Where(cond).
		OrderBy(
			// NOTE. for now. who cares about relevance.
			// db.Raw(data.ProductQueryWeight, qraw),
			"id DESC",
		)
	query = filterSort.UpdateQueryBuilder(query)

	var products []*data.Product
	paginate := cursor.UpdateQueryBuilder(query)
	if err := paginate.All(&products); err != nil {
		render.Respond(w, r, err)
		return
	}
	cursor.Update(products)
	render.RenderList(w, r, presenter.NewProductList(ctx, products))
}
