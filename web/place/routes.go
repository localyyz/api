package place

import (
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/web/post"

	"github.com/pressly/chi"
)

func Routes() http.Handler {
	r := chi.NewRouter()

	r.Get("/trending", ListTrendingPlaces)
	r.Route("/:placeID", func(r chi.Router) {
		r.Use(PlaceCtx)

		//r.Get("/", GetPlace)
		r.Mount("/posts", post.Routes())
	})

	return r
}
