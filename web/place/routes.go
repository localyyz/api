package place

import "github.com/pressly/chi"

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/nearby", ListNearby)
	r.Get("/recent", ListRecent)
	r.Get("/following", ListFollowing)

	r.Route("/:placeID", func(r chi.Router) {
		r.Use(PlaceCtx)

		r.Get("/", GetPlace)
		r.Get("/products", ListProduct)
		r.Post("/share", Share)

		r.Post("/follow", FollowPlace)
		r.Delete("/follow", UnfollowPlace)
	})

	return r
}
