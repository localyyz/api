package place

import (
	"bitbucket.org/moodie-app/moodie-api/web/shopify"
	"github.com/pressly/chi"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/nearby", Nearby)
	r.Get("/recent", Recent)
	r.Get("/following", ListFollowing)
	r.Get("/all", ListPlaces)

	r.Route("/manage", func(r chi.Router) {
		r.Get("/", ListManagable)
	})

	r.Route("/:placeID", func(r chi.Router) {
		r.Use(PlaceCtx)

		r.Mount("/shopify", shopify.Routes())

		r.Get("/", GetPlace)
		r.Get("/promos", ListPromo)
		r.Post("/follow", FollowPlace)
		r.Delete("/follow", UnfollowPlace)
	})

	return r
}
