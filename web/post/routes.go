package post

import (
	"net/http"

	"github.com/pressly/chi"
)

func Routes() http.Handler {
	r := chi.NewRouter()

	r.Post("/", CreatePost)
	r.Route("/:postID", func(r chi.Router) {
		r.Use(PostCtx)

		r.Get("/", GetPost)
		r.Put("/", UpdatePost)
		r.Delete("/", DeletePost)
	})

	return r
}
