package place

import "github.com/pressly/chi"

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/trending", ListTrending)

	r.Group(func(r chi.Router) {
		r.Get("/nearby", NearbyPlaces)
		r.Post("/autocomplete", AutoCompletePlaces)
	})

	r.Route("/:placeID", func(r chi.Router) {
		r.Use(PlaceCtx)

		r.Get("/", GetPlace)

		r.Post("/posts", CreatePost)
		r.Get("/posts/recent", ListRecentPosts)
	})

	return r
}
