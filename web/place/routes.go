package place

import "github.com/pressly/chi"

func Routes() chi.Router {
	r := chi.NewRouter()

	// TODO: combine trending with nearby promotion endpoint
	//  sort by distance and trendings should be grouped by locale in the frontend...
	r.Get("/trending", Trending)
	r.Get("/nearby", Nearby)

	r.Route("/:placeID", func(r chi.Router) {
		r.Use(PlaceCtx)
		r.Get("/", GetPlace)
	})

	return r
}
