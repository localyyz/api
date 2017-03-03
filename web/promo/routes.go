package promo

import "github.com/pressly/chi"

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/history", ListHistory)
	r.Get("/active", ListActive)

	// return list of user managable promotions
	r.Route("/manage", func(r chi.Router) {
		r.Use(PromoManageCtx)

		r.Get("/", ListManagable)
		r.Post("/", CreatePromo)
		r.Post("/preview", PreviewPromo)
	})

	r.Route("/:promoID", func(r chi.Router) {
		r.Use(PromoCtx)

		r.Get("/", GetPromo)
		r.Get("/claims", GetClaims)

		r.Group(func(r chi.Router) {
			r.Use(ClaimCtx)
			r.Post("/claim", ClaimPromo)
			r.Delete("/save", UnSavePromo)
		})
	})

	return r
}
