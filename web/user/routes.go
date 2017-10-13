package user

import "github.com/go-chi/chi"

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Route("/me", func(r chi.Router) {
		r.Use(MeCtx)

		r.Get("/", GetUser)
		// Pong.
		r.Get("/ping", Ping)

		r.Put("/device", SetDeviceToken)
		r.Route("/address", func(r chi.Router) {
			r.Post("/", CreateAddress)
			r.Get("/", ListAddresses)
		})
	})

	return r
}
