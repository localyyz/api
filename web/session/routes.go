package session

import "github.com/pressly/chi"

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Delete("/", Logout)
	r.Post("/heartbeat", Heartbeat)

	return r
}
