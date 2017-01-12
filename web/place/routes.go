package place

import "github.com/pressly/chi"

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/nearby", Nearby)
	r.Get("/recent", Recent)
	r.Get("/following", ListFollowing)
	r.Get("/all", ListPlaces)
	r.Post("/autocomplete", AutoComplete)

	r.Route("/:placeID", func(r chi.Router) {
		r.Use(PlaceCtx)
		r.Get("/", GetPlace)
		r.Get("/promos", ListPromo)

		r.Post("/follow", FollowPlace)
		r.Delete("/follow", UnfollowPlace)
	})

	return r
}
