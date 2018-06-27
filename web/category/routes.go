package category

import (
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/chi"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", List)
	r.Route("/{categoryType}", func(r chi.Router) {
		r.Use(CategoryTypeCtx)
		r.Get("/", GetCategory)

		r.Route("/{subcategory}", func(r chi.Router) {
			r.Use(SubcategoryCtx)
			r.Route("/products", api.FilterRoutes(ListProducts))
		})
		r.Route("/products", api.FilterRoutes(ListProducts))
	})

	return r
}
