package checkout

import "github.com/pressly/chi"

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", CreateCheckout)
	r.Put("/", UpdateCheckout)

	return r
}
