package user

import "github.com/pressly/chi"

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Route("/me", func(r chi.Router) {
		r.Use(MeCtx)

		r.Get("/cart", GetCart)
		r.Put("/device", SetDeviceToken)
		r.Post("/nda", AcceptNDA)
	})

	return r
}
