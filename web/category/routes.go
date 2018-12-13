package category

import (
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/chi"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	// parse gender context
	r.With(api.FilterSortCtx).With(CategoryRootCtx).Get("/", List)

	// TODO: deprecate
	r.With(discountCtx(0.1, 1)).Route("/10/products", api.FilterRoutes(ListDiscountProducts))
	r.With(discountCtx(0.20, 0.49)).Route("/11/products", api.FilterRoutes(ListDiscountProducts))
	r.With(discountCtx(0.50, 0.69)).Route("/12/products", api.FilterRoutes(ListDiscountProducts))
	r.With(discountCtx(0.70, 1)).Route("/13/products", api.FilterRoutes(ListDiscountProducts))
	r.With(discountCtx(0.50, 1)).Route("/onsale", api.FilterRoutes(ListDiscountProducts))

	r.Post("/styles", ListStyles)
	r.Post("/merchants", ListMerchants)
	r.Mount("/{categoryID}", categoryRoutes())
	return r
}

func categoryRoutes() chi.Router {
	r := chi.NewRouter()

	r.Use(CategoryCtx)
	r.Get("/", GetCategory)
	r.Route("/products", api.FilterRoutes(ListProducts))
	r.Get("/merchants", ListMerchants)

	return r
}
