package shipping

import "github.com/pressly/chi"

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", ListShippingRates)
	r.Put("/", UpdateShippingMethod)

	return r
}
