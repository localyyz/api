package webhooks

import (
	"context"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/lib/sync"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/render"
)

type productListingWrapper struct {
	ProductListing *shopify.ProductList `json:"product_listing"`
}

func (w *productListingWrapper) Bind(r *http.Request) error {
	return nil
}

func ProductListingHandler(r *http.Request) (err error) {
	wrapper := new(productListingWrapper)
	if err := render.Bind(r, wrapper); err != nil {
		return api.ErrInvalidRequest(err)
	}
	defer r.Body.Close()

	ctx := r.Context()
	ctx = context.WithValue(ctx, "sync.list", []*shopify.ProductList{wrapper.ProductListing})

	switch shopify.Topic(ctx.Value("sync.topic").(string)) {
	case shopify.TopicProductListingsAdd:
		return sync.ShopifyProductListingsCreate(ctx)
	case shopify.TopicProductListingsUpdate:
		return sync.ShopifyProductListingsUpdate(ctx)
	case shopify.TopicProductListingsRemove:
		return sync.ShopifyProductListingsRemove(ctx)
	}

	return
}
