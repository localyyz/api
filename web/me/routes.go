package me

import (
	"net/http"

	"github.com/pressly/chi"
)

func Routes() http.Handler {
	r := chi.NewRouter()

	r.Get("/", GetMe)
	r.Get("/points", GetPoints)

	return r
}
