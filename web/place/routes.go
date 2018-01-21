package place

import (
	"bitbucket.org/moodie-app/moodie-api/web/category"
	"github.com/go-chi/chi"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/following", ListFollowing)
	r.Post("/approval", HandleApproval)

	r.Route("/{placeID}", func(r chi.Router) {
		r.Use(PlaceCtx)

		r.Mount("/categories", category.Routes())

		r.Get("/", GetPlace)
		r.Get("/products", ListProduct)

		r.Post("/share", Share)
		r.Post("/follow", FollowPlace)
		r.Delete("/follow", UnfollowPlace)
	})

	return r
}
