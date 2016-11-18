package promo

import "github.com/pressly/chi"

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/claimed", ListClaimed)
	r.Route("/:promoID", func(r chi.Router) {
		r.Use(PromoCtx)

		r.Get("/", GetPromo)
		r.Get("/claims", GetClaims)

		r.Route("/:action", func(r chi.Router) {
			r.Use(PromoActionCtx)
			r.Post("/", DoPromo)
		})
	})

	return r
}
