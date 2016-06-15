package web

import (
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/web/auth"
	"bitbucket.org/moodie-app/moodie-api/web/post"
	"bitbucket.org/moodie-app/moodie-api/web/user"

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
