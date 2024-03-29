package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/apparelsorter"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
	"upper.io/db.v3/lib/sqlbuilder"
)

type FilterSort struct {
	Sort    *Sort
	Filters []*Filter

	// internals
	filterBy db.RawValue
	selector sqlbuilder.Selector
	r        *http.Request
	w        http.ResponseWriter

	ctx context.Context
}

const FilterDelim = "|"

var (
	SortParam   = `sort`
	FilterParam = `filter`

	FilterSortCtxKey = `filter.sort`

	defaultSortField        = ""
	ErrInvalidFilterSortKey = errors.New("invalid filter or sort key")

	MaxPrice = 300
)

const (
	SortDesc = "desc"
	SortAsc  = "asc"
)

type Sort struct {
	Type      string `json:"type"`
	Direction string `json:"direction"`
}

type Filter struct {
	// TODO: expand type and use it to derive sql queries
	Type     string      `json:"type"`
	MinValue interface{} `json:"min"`
	MaxValue interface{} `json:"max"`
	Value    interface{} `json:"val"`
}

func (o *FilterSort) Write(b []byte) (int, error) {
	// am i doing something here?
	if o.filterBy != nil && !o.filterBy.Empty() {
		val, err := o.GetValues(o.r.Context())
		if err != nil {
			render.Respond(o.w, o.r, err)
			return 0, nil
		}
		render.Respond(o.w, o.r, val)
		return 0, nil
	}
	return o.w.Write(b)
}

func (o *FilterSort) WriteHeader(statusCode int) {
	o.w.WriteHeader(statusCode)
}

func (o *FilterSort) Header() http.Header {
	return o.w.Header()
}

func FilterSortHijacksCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		filterSort := NewFilterSort(w, r)
		ctx := context.WithValue(r.Context(), FilterSortCtxKey, filterSort)
		// filter sort hijacks the response
		// typically, the next line looks something like
		//
		// next.ServeHTTP(w, r.WithContext(ctx))
		// howerver we've hijacked the default http writer with our own
		next.ServeHTTP(filterSort, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func FilterSortCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		filterSort := NewFilterSort(w, r)
		ctx := context.WithValue(r.Context(), FilterSortCtxKey, filterSort)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func WithFilterBy(val interface{}) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			filterSort, ok := ctx.Value(FilterSortCtxKey).(*FilterSort)
			if !ok {
				filterSort = NewFilterSort(w, r)
			}
			filterSort.filterBy = db.Raw(val.(string))
			ctx = context.WithValue(ctx, FilterSortCtxKey, filterSort)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}

// wrapFilterRoutes injects filter value routes into the mux
// along with the origional handler function
// usage:
//
//   r.Route("/products", api.FilterRoutes(ListProduct))
//
func wrapFilterRoutes(r chi.Router, handlerFn http.HandlerFunc) {
	r.Use(FilterSortHijacksCtx)
	r.Handle("/", handlerFn)
	r.With(WithFilterBy("lower(p.brand)")).Handle("/brands", handlerFn)
	r.With(WithFilterBy("lower(pv.etc->>'size')")).Handle("/sizes", handlerFn)
	r.With(WithFilterBy("lower(pv.etc->>'color')")).Handle("/colors", handlerFn)
	r.With(WithFilterBy("lower(p.category->>'value')")).Handle("/subcategories", handlerFn)
	r.With(WithFilterBy("lower(p.category->>'type')")).Handle("/categories", handlerFn)
	r.With(WithFilterBy("pv.price")).Handle("/prices", handlerFn)
	r.With(WithFilterBy("lower(pl.name)")).Handle("/stores", handlerFn)
}

// FilterRoutes wraps an http handlerfunc with filter routes
// usage:
//
//   r.Route("/products", api.FilterRoutes(ListProduct))
//
func FilterRoutes(handlerFn http.HandlerFunc) func(r chi.Router) {
	return func(r chi.Router) {
		wrapFilterRoutes(r, handlerFn)
	}
}

func WithFilterRoutes(handlerFn http.HandlerFunc) chi.Router {
	r := chi.NewRouter()
	wrapFilterRoutes(r, handlerFn)
	return r
}

func NewFilterSort(w http.ResponseWriter, r *http.Request) *FilterSort {
	if r == nil {
		return &FilterSort{w: w, r: r}
	}

	ctx := r.Context()
	q := r.URL.Query()
	o := &FilterSort{w: w, r: r, ctx: ctx}
	if value := q.Get(SortParam); len(value) > 0 && value != defaultSortField {
		value, _ = url.QueryUnescape(value)
		if strings.HasPrefix(value, "-") || strings.HasSuffix(value, "desc") {
			o.Sort = &Sort{
				Type:      value[1:],
				Direction: SortDesc,
			}
		} else {
			o.Sort = &Sort{
				Type:      value,
				Direction: SortAsc,
			}
		}
	}

	for _, value := range q[FilterParam] {
		value, _ = url.QueryUnescape(value)
		f := &Filter{}
		for _, p := range strings.SplitN(value, ",", 2) {
			if strings.HasPrefix(p, "min") {
				f.MinValue = p[4:]
			} else if strings.HasPrefix(p, "max") {
				f.MaxValue = p[4:]
			} else if strings.HasPrefix(p, "val") {
				f.Value = strings.TrimSpace(p[4:])
			} else {
				f.Type = p
			}
		}
		o.Filters = append(o.Filters, f)
	}

	if len(o.Filters) > 0 {
		lg.SetEntryField(r.Context(), "filter", o.Filters)
	}
	if o.Sort != nil {
		lg.SetEntryField(r.Context(), "sort", *o.Sort)
	}

	return o
}

func (o *FilterSort) Gender() *Filter {
	for _, f := range o.Filters {
		if f.Type == "gender" {
			return f
		}
	}
	return nil
}

func (o *FilterSort) HasFilter() bool {
	return o.filterBy != nil && !o.filterBy.Empty()
}

func (o *FilterSort) GetValues(ctx context.Context) ([]string, error) {
	var rows *sql.Rows
	var err error
	if strings.Contains(o.filterBy.Raw(), "pl.name") {
		rows, err = data.DB.Select(o.filterBy).
			From("places pl").
			Where(
				db.Raw(
					"pl.id IN (?)",
					o.selector.SetColumns(db.Raw("p.place_id")).
						GroupBy("p.place_id").
						OrderBy(nil),
				),
			).
			Query()
	} else if strings.Contains(o.filterBy.Raw(), "pv.price") {
		row, err := data.DB.Select(
			db.Raw("min(pv.price)"),
			db.Raw("max(pv.price)"),
		).
			From("product_variants pv").
			Where(
				db.Raw(
					"product_id IN (?)",
					o.selector.
						SetColumns("p.id").
						OrderBy("p.score DESC").
						Limit(100),
				),
				db.Cond{"pv.price": db.Gt(0)},
			).
			QueryRow()
		if err != nil {
			return []string{}, err
		}

		var min float64
		var max float64
		if err := row.Scan(&min, &max); err != nil {
			return []string{}, err
		}
		return []string{
			fmt.Sprintf("%d", int64(min)),
			fmt.Sprintf("%d", int64(max)),
		}, nil
	} else if strings.Contains(o.filterBy.Raw(), "pv.") {
		// TODO: really? best way? cache options on product_sizes / product_colors child table?
		// some kind of quick look up that makes it easier to do this???
		// get top 100 products from the selector
		rows, err = data.DB.Select(o.filterBy).
			From("product_variants pv").
			Where(
				db.Raw(
					"product_id IN (?)",
					o.selector.
						SetColumns("p.id").
						OrderBy("p.score DESC").
						Limit(100),
				),
				db.Cond{o.filterBy: db.NotEq("")},
			).
			GroupBy(o.filterBy).
			OrderBy(db.Raw("count(1) DESC")).
			Limit(30).
			Query()
	} else if o.selector != nil {
		rows, err = o.selector.
			SetColumns(o.filterBy).
			Where(db.Cond{
				o.filterBy: db.NotEq(""),
			}).
			GroupBy(o.filterBy).
			OrderBy(db.Raw("count(1) DESC")).
			Limit(100).
			Query()
	} else {
		err = errors.New("no selection query setup")
	}

	if err != nil {
		return []string{}, err
	}

	values := []string{}
	for rows.Next() {
		var v string
		err = rows.Scan(&v)
		if err != nil {
			break
		}
		values = append(values, strings.TrimSpace(v))
	}
	rows.Close()

	// context aware sorter
	if len(values) > 0 {
		if strings.Contains(o.filterBy.Raw(), "size") {
			sizesorter := apparelsorter.New(values...)
			sort.Sort(sizesorter)
			values = sizesorter.StringSlice()
		} else {
			sort.Strings(values)
		}
	}

	return values, nil
}

func parseFilter(v interface{}) (s []string) {
	vv, ok := v.(string)
	if !ok || v == "" {
		return
	}
	return strings.Split(strings.ToLower(vv), FilterDelim)
}

func (o *FilterSort) UpdateQueryBuilder(selector sqlbuilder.Selector) sqlbuilder.Selector {
	if s := o.Sort; s != nil {
		var orderBy string
		switch s.Type {
		case "price":
			orderBy = fmt.Sprintf("p.price %s", s.Direction)
		case "discount":
			orderBy = "p.discount_pct DESC"
		case "created_at":
			orderBy = "p.created_at DESC"
		default:
			orderBy = ""
		}
		selector = selector.OrderBy(
			db.Raw(orderBy),
			"p.id desc",
		)
	}
	for _, f := range o.Filters {
		var fConds []db.Compound
		// by default. let's filter out score greater equal to 1
		// NOTE: this is to increase query performance. sorting by multiple
		// values will dramatically slow down the query
		switch f.Type {
		case "brand", "brands":
			// list of brands
			if brands := parseFilter(f.Value); len(brands) > 0 {
				fConds = append(fConds, db.Cond{
					db.Raw("lower(p.brand)"): db.In(brands),
				})
			}
		case "place_id", "place_ids":
			fConds = append(fConds, db.Cond{"p.place_id": f.Value})
		case "gender", "genders":
			var genders []data.ProductGender
			for _, g := range parseFilter(f.Value) {
				v := new(data.ProductGender)
				if err := v.UnmarshalText([]byte(g)); err != nil {
					continue
				}
				if *v != data.ProductGenderUnisex {
					genders = append(genders, *v)
				}
			}
			lg.Warn(genders)
			if len(genders) > 0 {
				fConds = append(fConds, db.Cond{"p.gender": genders})
			}
		case "categoryType", "categoryTypes":
			fConds = append(fConds, db.Cond{
				db.Raw("lower(p.category->>'type')"): strings.ToLower(f.Value.(string)),
			})
		case "categoryValue", "categoryValues":
			if vals := parseFilter(f.Value); len(vals) > 0 {
				fConds = append(fConds, db.Cond{
					db.Raw("lower(p.category->>'value')"): vals,
				})
			}
		case "categories":
			var ancIDs []int64
			if err := json.Unmarshal([]byte(f.Value.(string)), &ancIDs); err == nil {
				if len(ancIDs) > 0 {
					var catIDs []int64
					for _, ID := range ancIDs {
						descIDs, _ := data.DB.Category.FindDescendantIDs(ID)
						catIDs = append(catIDs, ID)
						catIDs = append(catIDs, descIDs...)
					}
					fConds = append(fConds, db.Cond{
						db.Raw("p.category_id"): catIDs,
					})
				}
			}
		case "size", "sizes":
			// TODO: clean this up? is there a better way?
			if sizes := parseFilter(f.Value); len(sizes) > 0 {
				sizeSelector := data.DB.
					Select("pv.product_id").
					From("product_variants pv").
					Where(db.Cond{
						db.Raw("lower(pv.etc->>'size')"): db.In(sizes),
					})
				selector = selector.Where(db.Cond{"p.id IN": sizeSelector})
			}
		case "color", "colors":
			// TODO: clean this up? is there a better way?
			if colors := parseFilter(f.Value); len(colors) > 0 {
				colorSelector := data.DB.
					Select("pv.product_id").
					From("product_variants pv").
					Where(db.Cond{
						db.Raw("lower(pv.etc->>'color')"): colors,
					})
				selector = selector.Where(db.Cond{"p.id IN": colorSelector})
			}
		case "merchant", "merchants":
			if merchants := parseFilter(f.Value); len(merchants) > 0 {
				merchantSelector := data.DB.
					Select("pl.id").
					From("places pl").
					Where(db.Cond{
						db.Raw("lower(pl.name)"): db.In(merchants),
					})
				selector = selector.Where(db.Cond{"p.place_id IN": merchantSelector})
			}
		case "discount", "discounts":
			minDiscountPct, _ := strconv.ParseFloat(f.MinValue.(string), 64)
			fConds = append(fConds, db.Cond{"p.discount_pct": db.Gte(minDiscountPct / 100.0)})
		case "price", "prices":
			if f.MinValue != nil {
				min, _ := strconv.ParseFloat(f.MinValue.(string), 64)
				fConds = append(fConds, db.Cond{"p.price": db.Gte(min)})
			}
			if f.MaxValue != nil {
				max, _ := strconv.ParseFloat(f.MaxValue.(string), 64)
				fConds = append(fConds, db.Cond{"p.price": db.Lte(max)})
			}
		case "score", "scores":
			if f.MinValue != nil {
				min, _ := strconv.Atoi(f.MinValue.(string))
				fConds = append(fConds, db.Cond{"p.score": db.Gte(min)})
			}
		case "personalize":
			prf := &data.UserPreference{}
			if err := json.Unmarshal([]byte(f.Value.(string)), prf); err != nil {
				continue
			}
			user, ok := o.ctx.Value("session.user").(*data.User)
			if !ok || user.Preference == nil {
				continue
			}
			userPref := *user.Preference
			if len(prf.Styles) > 0 {
				userPref.Styles = prf.Styles
			}
			if len(prf.Pricings) > 0 {
				userPref.Pricings = prf.Pricings
			}
			if len(prf.Gender) > 0 {
				userPref.Gender = prf.Gender
			}
			placeIDs, err := data.DB.PlaceMeta.GetPlacesFromPreference(&userPref)
			if err != nil {
				continue
			}
			fConds = append(fConds, db.Cond{"p.place_id": placeIDs})
		}

		selector = selector.Where(db.And(fConds...))
	}

	o.selector = selector
	return selector
}
