package product

import (
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"bitbucket.org/moodie-app/moodie-api/web/auth"
	"bitbucket.org/moodie-app/moodie-api/web/user"
	"github.com/go-chi/chi"
)

func Routes() chi.Router {

	r := api.WithFilterRoutes(ListProducts)
	r.Route("/feedv3", func(r chi.Router) {
		r.Use(api.FilterSortCtx)
		r.Get("/", ListFeedV3)
		r.Route("/onsale", api.FilterRoutes(ListFeedV3Onsale))
		r.Route("/products", api.FilterRoutes(ListFeedV3Products))
	})

	r.Get("/trend", ListTrending)
	r.Get("/history", ListHistoryProduct)
	r.With(auth.DeviceCtx).Route("/favourite", api.FilterRoutes(ListFavourite))
	r.Route("/{productID}", func(r chi.Router) {
		r.Use(ProductCtx)
		r.Get("/", GetProduct)
		r.Get("/variant", GetVariant)
		r.Route("/related", api.FilterRoutes(ListRelatedProduct))
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
