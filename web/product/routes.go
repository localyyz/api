package product

import (
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"bitbucket.org/moodie-app/moodie-api/web/auth"
	"github.com/go-chi/chi"
)

func Routes() chi.Router {

	r := api.WithFilterRoutes(ListProducts)
	r.Get("/history", ListHistoryProduct)
	r.Get("/trend", ListTrending)
	r.Route("/favourite", api.FilterRoutes(ListFavourite))
	r.Route("/{productID}", func(r chi.Router) {
		r.Use(ProductCtx)
		r.Get("/", GetProduct)
		r.Get("/variant", GetVariant)
		r.Get("/related", ListRelatedProduct)
		r.Group(func(r chi.Router) {
			r.Use(auth.DeviceCtx)
			r.Post("/favourite", AddFavouriteProduct)
			r.Delete("/favourite", DeleteFavouriteProduct)
		})
	})

	return r
}
