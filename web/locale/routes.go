package locale

import "github.com/pressly/chi"

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", ListLocales)

	return r

}
