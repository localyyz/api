package shopify

import (
	"github.com/go-chi/chi"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	// TODO: is this terrible? Probably.
	// make a caching layer so we can accept
	// all the connection.
	r.Use(ShopifyStoreWhCtx)

	r.Post("/", WebhookHandler)
	return r
}
