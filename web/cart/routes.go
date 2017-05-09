package cart

import "github.com/pressly/chi"

func Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", CreateCart)

	return r
}
