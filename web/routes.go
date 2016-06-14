package web

import (
	"net/http"

	"bitbucket.org/pxue/api/web/auth"
	"bitbucket.org/pxue/api/web/me"
	"bitbucket.org/pxue/api/web/post"

	"github.com/pressly/chi"
)

func New() http.Handler {
	r := chi.NewRouter()

	r.Post("/login/facebook", auth.FacebookLogin)

	r.Group(func(r chi.Router) {
		r.Use(auth.SessionCtx)
		r.Route("/session", func(r chi.Router) {
			r.Delete("/logout", auth.Logout)
		})

		r.Mount("/me", me.Routes())
		r.Mount("/posts", post.Routes())
	})

	return r
}
