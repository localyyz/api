package locale

import "github.com/pressly/chi"

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", ListLocale)
	r.Route("/:localeID", func(r chi.Router) {
		r.Use(LocaleCtx)
		r.Get("/places", ListPlaces)
	})

	return r
}