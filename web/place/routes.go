package place

import (
	"net/http"

	"github.com/pressly/chi"
)

func Routes() http.Handler {
	r := chi.NewRouter()

	r.Get("/trending", ListTrendingPlaces)
	r.Route("/:placeID", func(r chi.Router) {
		//r.Use(PlaceCtx)

		//r.Get("/", GetPlace)
	})

	return r
}
