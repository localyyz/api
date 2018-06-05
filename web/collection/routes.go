package collection

import (
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/chi"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", ListCollection)
	r.Route("/{collectionID}", func(r chi.Router) {
		r.Use(CollectionCtx)
		r.Use(api.FilterSortCtx)
		r.Get("/", GetCollection)
		r.Get("/products", ListProducts)
	})

	return r
}
