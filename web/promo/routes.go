package promo

import "github.com/pressly/chi"

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/history", ListHistory)
	r.Get("/active", ListActive)
	r.Route("/:promoID", func(r chi.Router) {
		r.Use(PromoCtx)

		r.Get("/", GetPromo)
		r.Get("/claims", GetClaims)

		r.Group(func(r chi.Router) {
			r.Use(ClaimCtx)
			r.Post("/claim", ClaimPromo)
			r.Post("/save", SavePromo)
			r.Delete("/save", UnSavePromo)
		})
	})

	return r
}