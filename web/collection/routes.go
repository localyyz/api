package collection

import (
	"github.com/go-chi/chi"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", ListCollection)

	r.Route("/{collectionID}", func(r chi.Router) {
		r.Use(CollectionCtx)
		r.Use(CollectionProductCtx)
		r.Get("/", GetCollection)

		r.Get("/categories", ListCollectionCategory)
		r.Get("/brands", ListCollectionBrands)
		r.Get("/colors", ListCollectionColors)
		r.Get("/sizes", ListCollectionSizes)

		r.Get("/products", ListCollectionProduct)
	})

	return r
}
