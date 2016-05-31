package web

import (
	"net/http"

	"bitbucket.org/pxue/api/web/auth"

	"github.com/pressly/chi"
)

func New() http.Handler {
	r := chi.NewRouter()

	r.Post("/login/:network", auth.AuthCtx, auth.Login)

	return r
}
