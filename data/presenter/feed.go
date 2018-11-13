package presenter

import (
	"context"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/go-chi/render"
)

type Feed struct {
	*data.Feed

	Products []*Product `json:"products"`
	ctx      context.Context
}

func NewFeed(ctx context.Context, feed *data.Feed) *Feed {
	return &Feed{
		Feed:     feed,
		Products: newProductList(ctx, feed.Products),
		ctx:      ctx,
	}
}

func NewFeedList(ctx context.Context, feeds []*data.Feed) []render.Renderer {
	list := []render.Renderer{}
	for _, f := range feeds {
		list = append(list, NewFeed(ctx, f))
	}
	return list
}

func (f *Feed) Render(w http.ResponseWriter, r *http.Request) error {
	// we have to force each product to render manually. because renderer does
	// not iterate renderer slide unless it's done with RenderList
	for _, p := range f.Products {
		p.Render(w, r)
	}
	return nil
}
