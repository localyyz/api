package place

import "github.com/pressly/chi"

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/trending", ListTrending)
	// TODO: combine trending with nearby promotion endpoint
	//  sort by distance and trendings should be grouped by locale in the
	//  frontend...

	r.Get("/nearby", NearbyPlaces)
	r.Post("/autocomplete", AutoCompletePlaces)

	r.Route("/:placeID", func(r chi.Router) {
		r.Use(PlaceCtx)

		r.Get("/", GetPlace)
		r.Post("/posts", CreatePost)
		r.Post("/peek", PeekPromo)
		r.Get("/posts/recent", ListRecentPosts)
	})

	return r
}
