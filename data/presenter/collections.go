package presenter

import (
	"context"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
)

type Collection struct {
	*data.Collection
	Products []*Product `json:"products"`

	// metadata
	ProductCount uint64  `json:"productCount"`
	TotalSavings float64 `json:"totalSavings"`

	ctx context.Context
}

func NewCollection(ctx context.Context, collection *data.Collection) *Collection {
	c := &Collection{
		Collection: collection,
		ctx:        ctx,
	}

	cps, _ := data.DB.CollectionProduct.FindByCollectionID(c.ID)
	cpsIDs := make([]int64, len(cps))
	for i, ci := range cps {
		cpsIDs[i] = ci.ProductID
	}
	c.ProductCount = uint64(len(cps))

	if len(cpsIDs) > 0 {
		row, err := data.DB.Select(db.Raw("sum((price/(1-discount_pct))-price) as total_savings")).
			From("products").
			Where(db.Cond{"id": cpsIDs}).
			QueryRow()
		if err != nil {
			lg.Warn(err, "query collection saving")
			return c
		}
		if err := row.Scan(&c.TotalSavings); err != nil {
			lg.Warn(err, "present collection saving")
			return c
		}
	}

	return c
}

func (c *Collection) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func NewCollectionList(ctx context.Context, collections []*data.Collection) []render.Renderer {
	list := []render.Renderer{}
	for _, collection := range collections {
		c := NewCollection(ctx, collection)
		if c.ProductCount == 0 && c.PlaceIDs != nil && c.Categories != nil {
			continue
		}
		list = append(list, c)
	}
	return list
}
