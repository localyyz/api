package presenter

import (
	"context"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/go-chi/render"
	db "upper.io/db.v3"
)

type Collection struct {
	*data.Collection
	ProductCount uint64     `json:"productCount"`
	Products     []*Product `json:"products"`
	ctx          context.Context
}

func NewCollection(ctx context.Context, collection *data.Collection) *Collection {
	c := &Collection{
		Collection: collection,
		ctx:        ctx,
	}
	c.ProductCount, _ = data.DB.CollectionProduct.Find(db.Cond{"collection_id": c.ID}).Count()

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
