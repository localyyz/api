package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
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
}

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
			filterSort.filterBy = db.Raw(fmt.Sprintf("lower(%s)", val.(string)))
			ctx = context.WithValue(ctx, FilterSortCtxKey, filterSort)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}

func wrapFilterRoutes(r chi.Router, handlerFn http.HandlerFunc) {
	r.Use(FilterSortHijacksCtx)
	r.Get("/", handlerFn)
	r.With(WithFilterBy("p.brand")).Get("/brands", handlerFn)
	r.With(WithFilterBy("pv.etc->>'size'")).Get("/sizes", handlerFn)
	r.With(WithFilterBy("pv.etc->>'color'")).Get("/colors", handlerFn)
	r.With(WithFilterBy("p.category->>'value'")).Get("/subcategories", handlerFn)
	r.With(WithFilterBy("p.category->>'type'")).Get("/categories", handlerFn)
}

// TODO: turn these into middlewares?
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

	q := r.URL.Query()
	o := &FilterSort{w: w, r: r}
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
		for _, p := range strings.Split(value, ",") {
			if strings.HasPrefix(p, "min") {
				f.MinValue = p[4:]
			} else if strings.HasPrefix(p, "max") {
				f.MaxValue = p[4:]
			} else if strings.HasPrefix(p, "val") {
				f.Value = p[4:]
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

func (o *FilterSort) HasFilter() bool {
	return o.filterBy != nil && !o.filterBy.Empty()
}

func (o *FilterSort) GetValues(ctx context.Context) ([]string, error) {
	var rows *sql.Rows
	var err error
	if strings.Contains(o.filterBy.Raw(), "pv.") {
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
						OrderBy("score DESC").
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
			Limit(50).
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
		values = append(values, v)
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
			orderBy = "p.score DESC"
		}
		selector = selector.OrderBy(
			db.Raw(orderBy),
			"p.score desc",
			"p.id desc",
		)
	}
	for _, f := range o.Filters {
		var fConds []db.Compound
		switch f.Type {
		case "discount":
			fConds = append(fConds, db.Cond{"p.discount_pct": db.Gte(f.MinValue)})
		case "brand":
			fConds = append(fConds, db.Cond{
				db.Raw("lower(p.brand)"): strings.ToLower(f.Value.(string)),
			})
		case "place_id":
			fConds = append(fConds, db.Cond{"p.place_id": f.Value})
		case "gender":
			v := new(data.ProductGender)
			if err := v.UnmarshalText([]byte(f.Value.(string))); err != nil {
				lg.Warn(err)
				continue
			}
			fConds = append(fConds, db.Cond{"p.gender": *v})
		case "categoryType":
			fConds = append(fConds, db.Cond{
				db.Raw("lower(p.category->>'type')"): strings.ToLower(f.Value.(string)),
			})
		case "categoryValue":
			fConds = append(fConds, db.Cond{
				db.Raw("lower(p.category->>'value')"): strings.ToLower(f.Value.(string)),
			})
		case "size":
			// TODO: clean this up? is there a better way?
			sizeSelector := data.DB.
				Select("pv.product_id").
				From("product_variants pv").
				Where(db.Cond{
					db.Raw("lower(pv.etc->>'size')"): strings.ToLower(f.Value.(string)),
				})
			selector = selector.Where(db.Cond{"p.id IN": sizeSelector})
		case "color":
			// TODO: clean this up? is there a better way?
			colorSelector := data.DB.
				Select("pv.product_id").
				From("product_variants pv").
				Where(db.Cond{
					db.Raw("lower(pv.etc->>'color')"): strings.ToLower(f.Value.(string)),
				})
			selector = selector.Where(db.Cond{"p.id IN": colorSelector})
		case "price":
			if f.MinValue != nil {
				fConds = append(fConds, db.Cond{"p.price": db.Gte(f.MinValue)})
			}
			if f.MaxValue != nil {
				//the frontend can only return 300 as the max so anything above that there should be no Lte
				if f.MaxValue != MaxPrice {
					fConds = append(fConds, db.Cond{"p.price": db.Lte(f.MaxValue)})
				}
			}
		}
		selector = selector.Where(db.And(fConds...))
	}

	o.selector = selector
	return selector
}
