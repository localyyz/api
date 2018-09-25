package user

import (
	"github.com/go-chi/chi"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Route("/me", func(r chi.Router) {
		r.Use(MeCtx)

		r.Get("/", GetUser)
		// Pong.
		r.Get("/ping", Ping)

		r.Put("/", UpdateUser)
		r.Mount("/address", addressRoutes())
		r.Mount("/orders", orderRoutes())
	})

	r.Mount("/collections", collectionRoutes())

	return r
}

func orderRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", ListOrders)
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

func collectionRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", ListUserCollections)
	r.Post("/", CreateUserCollection)

	r.Route("/{collectionID}", func(r chi.Router) {
		r.Use(UserCollectionCtx)

		r.Get("/", GetUserCollection)
		r.Put("/", UpdateUserCollection)
		r.Delete("/", DeleteUserCollection)

		r.Route("/products", func(r chi.Router) {
			r.Get("/", GetUserCollectionProducts)
			r.Post("/", CreateProductInCollection)
		})
	})

	return r
}
