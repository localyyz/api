package product

import (
	"github.com/go-chi/chi"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/recent", ListRecentProduct)
	r.Get("/featured", ListFeaturedProducts)
	r.Route("/{productID}", func(r chi.Router) {
		r.Use(ProductCtx)
		r.Get("/", GetProduct)
		r.Get("/variant", GetVariant)
		r.Get("/related", ListRelatedProduct)
	})

	return r
}
