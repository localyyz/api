package collection

import (
	"github.com/go-chi/chi"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", ListCollection)

	r.Route("/{collectionID}", func(r chi.Router) {
		r.Use(CollectionCtx)
		r.Get("/", GetCollection)
		r.Get("/products", GetCollectionProduct)
	})

	return r
}
