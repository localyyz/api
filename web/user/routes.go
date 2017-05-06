package user

import "github.com/pressly/chi"

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Route("/me", func(r chi.Router) {
		r.Use(MeCtx)
		r.Mount("/", UserRoutes())
		r.Post("/nda", AcceptNDA)
	})

	r.Route("/:userID", func(r chi.Router) {
		r.Use(UserCtx)
		r.Mount("/", UserRoutes())
	})

	return r
}

func UserRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", GetUser)
	r.Get("/cart", GetCart)
	r.Put("/device", SetDeviceToken) // me

	return r
}
