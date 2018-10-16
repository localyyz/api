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

	// metadata
	ProductCount  uint64  `json:"productCount"`
	Collaborators []*User `json:"collaborators"`
	// TODO preview -> return some count of products

	ctx context.Context
}

func NewCollection(ctx context.Context, collection *data.Collection) *Collection {
	c := &Collection{
		Collection: collection,
		ctx:        ctx,
	}

	// count number of products
	c.ProductCount, _ = data.DB.CollectionProduct.Find(db.Cond{"collection_id": c.ID}).Count()

	if c.OwnerID != nil {
		if owner, _ := data.DB.User.FindByID(*c.OwnerID); owner != nil {
			c.Collaborators = []*User{{User: owner}}
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
		list = append(list, NewCollection(ctx, collection))
	}
	return list
}
