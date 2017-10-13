package checkout

import "github.com/go-chi/chi"

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", CreateCheckout)
	r.Put("/", UpdateCheckout)

	return r
}
