package cart

import "github.com/pressly/chi"

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", CreateCart)

	r.Route("/:cartID", func(r chi.Router) {
		r.Use(CartCtx)
		r.Get("/", GetCart)
		r.Put("/", UpdateCart)
		r.Delete("/", DeleteCart)

		r.Post("/checkout", Checkout)
		r.Post("/payment", Payment)

		r.Mount("/items", cartItemRoutes())
		r.Mount("/shipping", shippingRoutes())
	})

	return r
}

func shippingRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", ListShippingMethods)
	r.Put("/", UpdateShippingMethod)

	return r
}

func cartItemRoutes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", CreateCartItem)
	r.Route("/:cartItemID", func(r chi.Router) {
		r.Use(CartItemCtx)
		r.Get("/", GetCartItem)
		r.Put("/", UpdateCartItem)
		r.Delete("/", RemoveCartItem)
	})

	return r
}
