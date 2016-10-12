package place

import "github.com/pressly/chi"

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/nearby", Nearby)

	r.Route("/:placeID", func(r chi.Router) {
		r.Use(PlaceCtx)
		r.Get("/", GetPlace)
	})

	return r
}
