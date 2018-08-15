package category

import (
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/chi"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(api.FilterSortCtx)
	r.Get("/", List)
	r.With(discountCtx(0.1, 1)).Route("/sales/products", api.FilterRoutes(ListDiscountProducts))
	r.With(discountCtx(0.20, 0.5)).Route("/20% OFF/products", api.FilterRoutes(ListDiscountProducts))
	r.With(discountCtx(0.50, 0.7)).Route("/50% OFF/products", api.FilterRoutes(ListDiscountProducts))
	r.With(discountCtx(0.70, 1)).Route("/70% OFF/products", api.FilterRoutes(ListDiscountProducts))

	r.Route("/{categoryType}", func(r chi.Router) {
		r.Use(CategoryTypeCtx)
		r.Get("/", GetCategory)
		r.Route("/products", api.FilterRoutes(ListProducts))
	})
	return r
}
