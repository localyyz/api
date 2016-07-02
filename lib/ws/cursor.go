package ws

import (
	"fmt"
	"math"
	"net/http"
	"net/url"
	"reflect"
	"strconv"

	"upper.io/db"
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
	Limit      int
	NextPage   bool
	firstOnly  bool // Request to return only the first record, as a singular object.
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
	total, _ := res.Count()
	p.TotalPages = int(math.Ceil(float64(total) / float64(p.Limit)))
	if p.Page > 1 {
		return res.Limit(uint(p.Limit)).Skip(uint((p.Page - 1) * p.Limit))
	}
	return res.Limit(uint(p.Limit))
}

func (p *Page) HasFirst() bool { return true }

func (p *Page) HasLast() bool { return false }

func (p *Page) HasPrev() bool { return (p.Page > 0) }

func (p *Page) HasNext() bool { return p.NextPage }

func (p *Page) FirstOnly() bool { return p.firstOnly }
