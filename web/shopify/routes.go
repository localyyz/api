package shopify

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	// TODO: is this terrible? Probably.
	// make a caching layer so we can accept
	// all the connection.
	r.Use(middleware.Throttle(200))
	r.Use(ShopifyStoreWhCtx)

	r.Post("/", WebhookHandler)
	return r
}
