package user

import (
	"bitbucket.org/moodie-app/moodie-api/web/cart"
	"github.com/pressly/chi"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Route("/me", func(r chi.Router) {
		r.Use(MeCtx)

		r.Get("/", GetUser)
		// Pong.
		r.Get("/ping", Ping)

		r.Route("/carts/:scope", func(r chi.Router) {
			r.Use(cart.CartScopeCtx)
			r.Get("/", cart.ListCarts)
		})
		r.Get("/carts", cart.ListCarts)

		r.Put("/device", SetDeviceToken)

		r.Route("/address", func(r chi.Router) {
			r.Post("/", CreateAddress)
			r.Get("/", ListAddresses)
		})

		r.Mount("/payments", paymentRoutes())
	})

	return r
}

func paymentRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", ListPaymentMethods)
	r.Post("/", CreatePaymentMethod)
	r.Route("/:paymentID", func(r chi.Router) {
		r.Use(PaymentMethodCtx)
		r.Get("/", GetPaymentMethod)
		r.Put("/", UpdatePaymentMethod)
		r.Delete("/", RemovePaymentMethod)
	})

	return r
}
