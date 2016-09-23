package promo

import "github.com/pressly/chi"

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Route("/:promoID", func(r chi.Router) {
		r.Use(PromoCtx)

		r.Get("/", GetPromo)
		r.Post("/claim", ClaimPromo)
	})

	return r
}
