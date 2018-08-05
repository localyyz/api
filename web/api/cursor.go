package api

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"time"

	"github.com/pressly/lg"
	"upper.io/db.v3"
	"upper.io/db.v3/lib/sqlbuilder"
)

var (
	PageParam            = `page`
	DefaultResourceLimit = 50
	MaxResourceLimit     = 100
	UntilParam           = `until`
	SinceParam           = `since`
	LimitParam           = `limit`
	DefaultKey           = `id`
)

type Page struct {
	URL        *url.URL
	Page       int
	TotalPages int
	ItemTotal  int
	Limit      int
	NextPage   bool
	firstOnly  bool // Request to return only the first record, as a singular object.
}

func PaginateCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		cursor := NewPage(r)
		ctx := context.WithValue(r.Context(), "cursor", cursor)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func NewPage(r *http.Request) *Page {
	if r == nil {
		return &Page{
			Page:  1,
			Limit: DefaultResourceLimit,
		}
	}

	// NOTE: Goji's SubRouter overwrites r.URL.
	u := r.URL

	page, _ := strconv.Atoi(u.Query().Get(PageParam))
	if page <= 0 {
		page = 1
	}

	limit, _ := strconv.Atoi(u.Query().Get(LimitParam))
	if limit <= 0 {
		limit = DefaultResourceLimit
	}
	if limit > MaxResourceLimit {
		limit = MaxResourceLimit
	}

	firstOnly := u.Query().Get("first") != ""
	if firstOnly {
		limit = 1
	}

	return &Page{
		URL:       u,
		Page:      page,
		Limit:     limit,
		firstOnly: firstOnly,
	}
}

func (p *Page) Update(v interface{}) *Page {
	if reflect.TypeOf(v).Kind() != reflect.Slice {
		p.NextPage = false
		return p
	}
	s := reflect.ValueOf(v)

	// TODO: this does not mean that there is a next page in all cases.
	p.NextPage = (s.Len() == p.Limit)

	return p
}

func (p *Page) PageURLs() map[string]string {
	links := map[string]string{}
	u := *p.URL
	q := u.Query()
	q.Set(LimitParam, fmt.Sprintf("%d", p.Limit))

	// First.
	q.Set(PageParam, "1")
	u.RawQuery = q.Encode()
	links["first"] = u.String()

	// Current.
	q.Set(PageParam, fmt.Sprintf("%d", p.Page))
	u.RawQuery = q.Encode()
	links["self"] = u.String()

	// Previous.
	if p.HasPrev() {
		q.Set(PageParam, fmt.Sprintf("%d", p.Page-1))
		u.RawQuery = q.Encode()
		links["prev"] = u.String()
	}

	// Next.
	if p.HasNext() {
		q.Set(PageParam, fmt.Sprintf("%d", p.Page+1))
		u.RawQuery = q.Encode()
		links["next"] = u.String()
	}

	return links
}

func (p *Page) DbCondition() db.Cond {
	return db.Cond{}
}

func (p *Page) UpdateQueryUpper(res db.Result) db.Result {
	{
		total, _ := res.Group(nil).Count()
		p.TotalPages = int(math.Ceil(float64(total) / float64(p.Limit)))
		p.ItemTotal = int(total)
	}
	if p.Page > 1 {
		return res.Limit(p.Limit).Offset((p.Page - 1) * p.Limit)
	}
	return res.Limit(p.Limit)
}

func (p *Page) UpdateQueryBuilder(selector sqlbuilder.Selector) sqlbuilder.Paginator {
	{
		t := 500 * time.Millisecond
		ctx, cancel := context.WithTimeout(context.Background(), t)
		defer cancel()

		done := make(chan int, 1)

		func() {
			//go func() {
			row, _ := selector.
				SetColumns(db.Raw("count(1)")).
				GroupBy(db.Raw("")).
				OrderBy(nil).
				QueryRowContext(ctx)
			var count uint64
			if err := row.Scan(&count); err != nil {
				lg.Println(err)
			}
			p.ItemTotal = int(count)
			done <- 1
		}()
		select {
		case <-done:
			// done!
		case <-time.After(t):
			fmt.Println("hard timeout")
		case <-ctx.Done():
			fmt.Println(ctx.Err()) // prints "context deadline exceeded"
		}
	}
	paginator := selector.Paginate(uint(p.Limit))
	if p.Page > 1 {
		return paginator.Page(uint(p.Page))
	}
	return paginator
}

func (p *Page) HasFirst() bool { return true }

func (p *Page) HasLast() bool { return false }

func (p *Page) HasPrev() bool { return (p.Page > 0) }

func (p *Page) HasNext() bool { return p.NextPage }

func (p *Page) FirstOnly() bool { return p.firstOnly }
