package place

import (
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/chi"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", List)                 // DEPRECATED
	r.Get("/featured", ListFeatured) // DEPRECATED
	r.Route("/{placeID}", func(r chi.Router) {
		r.Use(PlaceCtx)
		r.Get("/", GetPlace)
		r.Route("/products", api.FilterRoutes(ListProducts))
	})

	return r
}
