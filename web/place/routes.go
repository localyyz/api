package place

import "github.com/go-chi/chi"

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/nearby", ListNearby)
	r.Get("/recent", ListRecent)
	r.Get("/following", ListFollowing)

	r.Post("/approval", HandleApproval)

	r.Route("/{placeID}", func(r chi.Router) {
		r.Use(PlaceCtx)

		r.Get("/", GetPlace)
		r.Get("/products", ListProduct)

		// TODO: remove ~ legacy endpoint
		r.Get("/tags", ListProductTags)

		// filtering endpoints
		r.Get("/brands", ListProductBrands)
		r.Get("/categories", ListProductCategory)
		r.Get("/prices", ListProductPrices)

		r.Post("/share", Share)
		r.Post("/follow", FollowPlace)
		r.Delete("/follow", UnfollowPlace)
	})

	return r
}
