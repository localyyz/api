package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	db "upper.io/db.v3"
	"upper.io/db.v3/lib/sqlbuilder"
)

type FilterSort struct {
	Sort    *Sort
	Filters []*Filter
}

var (
	SortParam   = `sort`
	FilterParam = `filter`

	defaultSortField = ""

	ErrInvalidFilterSortKey = errors.New("invalid filter or sort key")
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
	Type     string      `json:"type"`
	MinValue interface{} `json:"min"`
	MaxValue interface{} `json:"max"`
	Value    interface{} `json:"val"`
}

func FilterSortCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		filterSort := NewFilterSort(r)
		ctx := context.WithValue(r.Context(), "filter.sort", filterSort)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func NewFilterSort(r *http.Request) *FilterSort {
	if r == nil {
		return &FilterSort{}
	}

	q := r.URL.Query()
	o := &FilterSort{}
	if value := q.Get(SortParam); len(value) > 0 && value != defaultSortField {
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

	return o
}

func (o *FilterSort) UpdateQueryBuilder(selector sqlbuilder.Selector) sqlbuilder.Selector {
	if o.Sort == nil && len(o.Filters) == 0 {
		return selector
	}

	if s := o.Sort; s != nil {
		var orderBy string
		switch s.Type {
		case "price":
			orderBy = fmt.Sprintf("p.price %s", s.Direction)
		case "discount":
			orderBy = "p.discount_pct DESC"
		case "created_at":
			orderBy = fmt.Sprintf("p.created_at %s", s.Direction)
		default:
			orderBy = "p.score DESC"
		}
		selector = selector.OrderBy(db.Raw(orderBy), "p.score desc", "p.id desc")
	}
	for _, f := range o.Filters {
		var fConds []db.Compound

		switch f.Type {
		case "discount":
			fConds = append(fConds, db.Cond{"discount_pct": db.Gte(f.MinValue)})
		case "brand":
			fConds = append(fConds, db.Cond{"brand": f.Value})
		case "place_id":
			fConds = append(fConds, db.Cond{"place_id": f.Value})
		case "gender":
			v := new(data.ProductGender)
			v.UnmarshalText([]byte(f.Value.(string)))
			fConds = append(fConds, db.Cond{"gender": v})
		default:
			if f.MinValue != nil {
				fConds = append(fConds, db.Cond{f.Type: db.Gte(f.MinValue)})
			}
			if f.MaxValue != nil {
				fConds = append(fConds, db.Cond{f.Type: db.Lte(f.MaxValue)})
			}
		}
		selector = selector.Where(db.And(fConds...))
	}
	return selector
}
