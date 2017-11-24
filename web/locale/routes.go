package locale

import "github.com/go-chi/chi"

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/cities", ListCities)
	r.Route("/{localeID}", func(r chi.Router) {
		r.Use(LocaleCtx)
		r.Get("/places", ListPlaces)
	})

	return r
}
