package search

import (
	"net/http"

	"github.com/pressly/chi"
)

func Routes() http.Handler {
	r := chi.NewRouter()

	r.Post("/places", AutocompletePlaces)

	return r
}
