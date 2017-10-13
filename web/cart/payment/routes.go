package payment

import "github.com/go-chi/chi"

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/methods", ListPaymentMethods)
	r.Post("/", CreatePayment)

	return r
}
