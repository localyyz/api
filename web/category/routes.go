package category

import (
	"github.com/go-chi/chi"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", ListCategory)

	r.Route("/{categoryType:[a-z]+}", func(r chi.Router) {
		r.Use(CategoryTypeCtx)
		r.Get("/", ListProductCategory)

		r.Get("/brands", ListProductBrands)
		r.Get("/colors", ListProductColors)
		r.Get("/sizes", ListProductSizes)

		r.Post("/products", ListCategoryProduct)
	})

	return r
}