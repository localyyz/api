package category

import (
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/chi"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.With(api.FilterSortCtx).With(CategoryRootCtx).Get("/", List)

	r.Post("/styles", ListStyles)
	r.Mount("/{categoryID}", categoryRoutes())

	return r
}

func categoryRoutes() chi.Router {
	r := chi.NewRouter()

	r.Use(CategoryCtx)
	r.Get("/", GetCategory)
	r.Route("/products", api.FilterRoutes(ListProducts))

	return r
}
