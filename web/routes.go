package web

import (
	"net/http"

	"bitbucket.org/pxue/api/web/auth"
	"bitbucket.org/pxue/api/web/post"
	"bitbucket.org/pxue/api/web/user"

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

		r.Mount("/users", user.Routes())
		r.Mount("/posts", post.Routes())
	})

	return r
}
