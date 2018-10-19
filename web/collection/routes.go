package collection

import (
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/chi"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/featured", ListFeaturedCollection)
	r.Route("/{collectionID}", func(r chi.Router) {
		r.Use(CollectionCtx)
		r.Get("/", GetCollection)
		r.Route("/products", api.FilterRoutes(ListProducts))
	})

	return r
}
