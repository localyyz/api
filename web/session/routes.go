package session

import "github.com/go-chi/chi"

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Delete("/", Logout)

	return r
}
