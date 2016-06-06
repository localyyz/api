package post

import (
	"net/http"

	"github.com/pressly/chi"
)

func Routes() http.Handler {
	r := chi.NewRouter()

	r.Post("/", CreatePost)
	r.Get("/trending", ListTrendingPost)
	r.Get("/fresh", ListFreshPost)
	r.Route("/:postID", func(r chi.Router) {
		r.Use(PostCtx)

		r.Mount("/like", LikeRoutes())
		r.Mount("/comment", CommentRoutes())

		r.Get("/", GetPost)
		r.Put("/", UpdatePost)    // user
		r.Delete("/", DeletePost) // user
	})

	return r
}

func LikeRoutes() http.Handler {
	r := chi.NewRouter()

	r.Post("/", LikePost)
	r.Get("/", ListPostLike)
	r.Route("/:likeID", func(r chi.Router) {
		r.Use(LikeCtx)

		r.Get("/", GetLike)
		r.Delete("/", UnlikePost)
	})

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
