package product

import (
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"bitbucket.org/moodie-app/moodie-api/web/auth"
	"bitbucket.org/moodie-app/moodie-api/web/user"
	"github.com/go-chi/chi"
)

func Routes() chi.Router {

	r := api.WithFilterRoutes(ListProducts)
	r.Route("/feed", api.FilterRoutes(ListRandomProduct))
	r.Route("/trend", api.FilterRoutes(ListTrending))
	r.Get("/history", ListHistoryProduct)
	r.With(auth.DeviceCtx).Route("/favourite", api.FilterRoutes(ListFavourite))
	r.Route("/{productID}", func(r chi.Router) {
		r.Use(ProductCtx)
		r.Get("/", GetProduct)
		r.Get("/variant", GetVariant)
		r.Get("/related", ListRelatedProduct)
		r.Group(func(r chi.Router) {
			r.Use(auth.DeviceCtx)
			r.Post("/favourite", AddFavouriteProduct)
			r.Delete("/favourite", DeleteFavouriteProduct)

			r.Route("/collections", func(r chi.Router) {
				r.Delete("/", DeleteFromAllCollections)
				r.Route("/{collectionID}", func(r chi.Router) {
					r.Use(user.UserCollectionCtx)
					r.Delete("/", DeleteProductFromCollection)
				})
			})
		})
	})

	return r
}
