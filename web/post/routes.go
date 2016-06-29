package post

import (
	"net/http"

	"github.com/pressly/chi"
)

func Routes() http.Handler {
	r := chi.NewRouter()

	r.Post("/", CreatePost)
	r.Get("/trending", ListTrendingPost)
	r.Get("/recent", ListFreshPost)
	r.Route("/:postID", func(r chi.Router) {
		r.Use(PostCtx)

		r.Mount("/likes", LikeRoutes())
		r.Mount("/comment", CommentRoutes())

		r.Get("/", GetPost)
		r.Put("/", UpdatePost)    // self user
		r.Delete("/", DeletePost) // self user
	})

	return r
}

func LikeRoutes() http.Handler {
	r := chi.NewRouter()

	r.Get("/", ListPostLike)
	r.Post("/", LikePost)
	r.Delete("/", UnlikePost)

	return r
}

func CommentRoutes() http.Handler {
	r := chi.NewRouter()

	r.Post("/", AddComment)
	r.Get("/", ListPostComment)
	r.Route("/:commentID", func(r chi.Router) {
		r.Use(CommentCtx)

		r.Get("/", GetComment)
		r.Delete("/", DeleteComment)
	})

	return r
}
