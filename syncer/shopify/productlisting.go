package shopify

import (
	"context"
	"net/http"

	lib "bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/lib/sync"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/render"
)

type productListingWrapper struct {
	ProductListing *lib.ProductList `json:"product_listing"`
}

func (w *productListingWrapper) Bind(r *http.Request) error {
	return nil
}

func ProductListingHandler(r *http.Request) (err error) {
	wrapper := new(productListingWrapper)
	if err := render.Bind(r, wrapper); err != nil {
		return api.ErrInvalidRequest(err)
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, "sync.list", []*lib.ProductList{wrapper.ProductListing})

	switch lib.Topic(ctx.Value("sync.topic").(string)) {
	case lib.TopicProductListingsAdd:
		return sync.ShopifyProductListingsCreate(ctx)
	case lib.TopicProductListingsUpdate:
		return sync.ShopifyProductListingsUpdate(ctx)
	case lib.TopicProductListingsRemove:
		return sync.ShopifyProductListingsRemove(ctx)
	}

	return
}
