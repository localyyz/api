package place

import (
	"bitbucket.org/moodie-app/moodie-api/web/post"

	"github.com/pressly/chi"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/search", SearchPlaces)
	r.Get("/trending", ListTrendingPlaces)
	r.Route("/:placeID", func(r chi.Router) {
		r.Use(PlaceCtx)

		r.Get("/", GetPlace)
		r.Mount("/posts", post.Routes())
	})

	return r
}
