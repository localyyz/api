package web

import (
	"net/http"

	"bitbucket.org/pxue/api/web/auth"

	"github.com/pressly/chi"
)

func New() http.Handler {
	r := chi.NewRouter()

	r.Post("/login/facebook", auth.FacebookLogin)

	r.Route("/session", func(r chi.Router) {
		r.Use(auth.SessionCtx)

		r.Delete("/logout", auth.Logout)
	})

	return r
}
