package presenter

import (
	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/stash"
	"context"
	"github.com/go-chi/render"
	"math"
	"net/http"
)

type CollectionUser struct {
	*data.UserCollection
	TotalProducts int64   `json:"productCount"`
	Savings       float64 `json:"savings"`
	Owner         *User   `json:"owner"`

	ctx context.Context
}

func (c *CollectionUser) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func NewUserCollection(ctx context.Context, collection *data.UserCollection) *CollectionUser {
	c := &CollectionUser{
		UserCollection: collection,
		ctx:            ctx,
	}

	// append the owner
	user := ctx.Value("session.user").(*data.User)
	c.Owner = NewUser(ctx, user)

	// get the product count and savings from redis
	c.TotalProducts, _ = stash.GetUserCollProdCount(collection.ID)

	// rounding to 2 decimal places
	savings, _ := stash.GetUserCollSavings(collection.ID)
	c.Savings = math.Round(savings*100) / 100

	return c
}

func NewUserCollectionList(ctx context.Context, collections []*data.UserCollection) []render.Renderer {
	list := []render.Renderer{}

	for _, collection := range collections {
		c := NewUserCollection(ctx, collection)
		list = append(list, c)
	}
	return list
}
