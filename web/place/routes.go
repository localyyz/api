package place

import (
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/chi"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", List)
	r.Get("/featured", ListFeatured)
	r.Route("/{placeID}", func(r chi.Router) {
		r.Use(PlaceCtx)
		r.Use(api.FilterSortCtx)

		r.Get("/", GetPlace)
		r.Get("/products", ListProducts)
	})

	return r
}
