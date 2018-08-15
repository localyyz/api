package category

import (
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/chi"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.With(api.FilterSortCtx).Get("/", List)
	r.With(api.FilterSortCtx).With(discountCtx(0.1, 1)).Get("/sales/products", ListDiscountProducts)
	r.With(api.FilterSortCtx).With(discountCtx(0.20, 0.5)).Get("/20% OFF/products", ListDiscountProducts)
	r.With(api.FilterSortCtx).With(discountCtx(0.50, 0.7)).Get("/50% OFF/products", ListDiscountProducts)
	r.With(api.FilterSortCtx).With(discountCtx(0.70, 1)).Get("/70% OFF/products", ListDiscountProducts)

	r.Route("/{categoryType}", func(r chi.Router) {
		r.Use(CategoryTypeCtx)
		r.Get("/", GetCategory)
		r.Route("/products", api.FilterRoutes(ListProducts))
	})
	return r
}
