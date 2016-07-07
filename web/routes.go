package web

import (
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/web/auth"
	"bitbucket.org/moodie-app/moodie-api/web/locale"
	"bitbucket.org/moodie-app/moodie-api/web/middleware/logger"
	"bitbucket.org/moodie-app/moodie-api/web/place"
	"bitbucket.org/moodie-app/moodie-api/web/post"
	"bitbucket.org/moodie-app/moodie-api/web/user"

	"github.com/pressly/chi"
	"github.com/pressly/chi/middleware"
)

func New() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.NoCache)
	r.Use(logger.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`¯\_(ツ)_/¯`))
	})

	r.Post("/login/facebook", auth.FacebookLogin)

	r.Group(func(r chi.Router) {
		r.Use(auth.SessionCtx)
		r.Route("/session", func(r chi.Router) {
			r.Delete("/logout", auth.Logout)
		})

		r.Mount("/users", user.Routes())
		r.Mount("/locale", locale.Routes())
		r.Mount("/places", place.Routes())
		r.Mount("/posts", post.Routes())
	})

	return r
}
