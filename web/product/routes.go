package product

import (
	"github.com/go-chi/chi"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/history", ListHistoryProduct)
	r.Get("/curated", ListCurated)
	r.Get("/onsale", ListOnsaleProduct)
	r.Route("/{productID}", func(r chi.Router) {
		r.Use(ProductCtx)
		r.Get("/", GetProduct)
		r.Get("/variant", GetVariant)
		r.Get("/related", ListRelatedProduct)
	})

	return r
}
