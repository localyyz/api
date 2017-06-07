package place

import (
	"bitbucket.org/moodie-app/moodie-api/web/shopify"
	"github.com/pressly/chi"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/nearby", ListNearby)
	r.Get("/recent", ListRecent)
	r.Get("/following", ListFollowing)

	r.Route("/:placeID", func(r chi.Router) {
		r.Use(PlaceCtx)

		r.Mount("/shopify", shopify.Routes())

		r.Get("/", GetPlace)
		r.Get("/promos", ListPromo)
		r.Post("/share", Share)

		r.Post("/follow", FollowPlace)
		r.Delete("/follow", UnfollowPlace)
	})

	return r
}
