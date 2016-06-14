package user

import (
	"net/http"

	"github.com/pressly/chi"
)

func Routes() http.Handler {
	r := chi.NewRouter()

	r.Mount("/me", MeCtx, UserRoutes())
	r.Mount("/:userId", UserCtx, UserRoutes())

	return r
}

func UserRoutes() http.Handler {
	r := chi.NewRouter()

	r.Get("/", GetUser)
	r.Get("/points", GetPointHistory) // self user
	r.Get("/posts/recent", GetRecentPost)

	return r
}
