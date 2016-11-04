package place

import "github.com/pressly/chi"

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/nearby", Nearby)
	r.Post("/search", Search)
	r.Post("/autocomplete", AutoComplete)

	r.Route("/:placeID", func(r chi.Router) {
		r.Use(PlaceCtx)
		r.Get("/", GetPlace)
	})

	return r
}
