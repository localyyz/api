package user

import "github.com/go-chi/chi"

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Route("/me", func(r chi.Router) {
		r.Use(MeCtx)

		r.Get("/", GetUser)
		// Pong.
		r.Get("/ping", Ping)

		r.Put("/", UpdateUser)
		r.Put("/device", SetDeviceToken)
		r.Mount("/address", addressRoutes())
	})

	return r
}

func addressRoutes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", CreateAddress)
	r.Get("/", ListAddresses)
	r.Route("/{addressID}", func(r chi.Router) {
		r.Use(AddressCtx)
		r.Get("/", GetAddress)
		r.Put("/", UpdateAddress)
		r.Delete("/", RemoveAddress)
	})

	return r
}
