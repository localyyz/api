package tool

import (
	"context"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/lib/sync"
	"github.com/pressly/lg"
)

func SyncCollections(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	client := ctx.Value("shopify.client").(*shopify.Client)
	place := ctx.Value("place").(*data.Place)

	page := 1
	for {
		params := &shopify.CollectionListParam{
			Page:  page,
			Limit: 1,
		}
		lists, _, err := client.CollectionList.List(ctx, params)
		if err != nil {
			lg.Alert(err)
			break
		}
		if len(lists) == 0 {
			lg.Info("done")
			break
		}
		for _, c := range lists {
			ctx = context.Background()
			ctx = context.WithValue(ctx, "shopify.client", client)
			ctx = context.WithValue(ctx, "sync.list", []*shopify.CollectionList{c})
			ctx = context.WithValue(ctx, "sync.place", place)
			if err := sync.ShopifyCollectionListingsUpdate(ctx); err != nil {
				lg.Warnf("failed to sync %s with %v", c.Title, err)
			}
		}
		page++
	}
}
