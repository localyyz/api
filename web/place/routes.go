package place

import (
	"bitbucket.org/moodie-app/moodie-api/web/post"
	"github.com/pressly/chi"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/trending", ListTrendingPlaces)

	r.Group(func(r chi.Router) {
		r.Use(PlaceTypeCtx)
		r.Get("/nearby", NearbyPlaces)
		r.Post("/search", SearchPlaces)
	})

	r.Route("/:placeID", func(r chi.Router) {
		r.Use(PlaceCtx)

		r.Get("/", GetPlace)
		r.Get("/posts/recent", post.ListFreshPost)
	})

	return r
}
