package data

import (
	"time"

	"bitbucket.org/moodie-app/moodie-api/lib/shopify"

	"upper.io/bond"
)

type WebhookCall struct {
	ID        int64           `db:"id,pk,omitempty"`
	Data      WebhookCallData `db:"data"`
	PlaceID   int64           `db:"place_id"`
	CreatedAt time.Time       `db:"created_at"`
}

type WebhookCallData struct {
	ProductListing *shopify.ProductList `db:"product_listing,omitempty"`
}

type WebhookCallStore struct {
	bond.Store
}

func (*WebhookCall) CollectionName() string {
	return `webhooks`
}
