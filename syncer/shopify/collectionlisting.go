package shopify

import (
	"context"
	"net/http"

	lib "bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/render"
)

type collectionListingWrapper struct {
	CollectionListing *lib.CollectionList `json:"collection_listings"`
}

func (w *collectionListingWrapper) Bind(r *http.Request) error {
	return nil
}

func CollectionListingHandler(r *http.Request) (err error) {
	wrapper := new(collectionListingWrapper)
	if err := render.Bind(r, wrapper); err != nil {
		return api.ErrInvalidRequest(err)
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, "sync.list", []*lib.CollectionList{wrapper.CollectionListing})

	switch lib.Topic(ctx.Value("sync.topic").(string)) {
	case lib.TopicCollectionListingsAdd:
		//return sync.ShopifyCollectionListingsCreate(ctx)
	case lib.TopicCollectionListingsUpdate:
		//return sync.ShopifyCollectionListingsUpdate(ctx)
	case lib.TopicCollectionListingsRemove:
		//return sync.ShopifyCollectionListingsRemove(ctx)
	}

	return
}
