package category

import (
	"bitbucket.org/moodie-app/moodie-api/web/product"
	"github.com/go-chi/chi"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.With(product.ProductGenderCtx).
		Get("/gender/{gender}", ListCategory)
	r.Route("/{categoryType:[a-z]+}", func(r chi.Router) {
		r.Use(CategoryTypeCtx)
		r.Get("/", GetCategory)

		r.Get("/brands", ListProductBrands)
		r.Get("/colors", ListProductColors)
		r.Get("/sizes", ListProductSizes)

		r.Post("/products", ListCategoryProduct)
	})

	return r
}
