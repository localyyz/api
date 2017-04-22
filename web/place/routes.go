package place

import (
	"net/http"

	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
	"github.com/pressly/chi"
)

func ConnectShopify(w http.ResponseWriter, r *http.Request) {
	// check if cred already exist
	place := r.Context().Value("place").(*data.Place)
	count, err := data.DB.ShopifyCred.Find(db.Cond{"place_id": place.ID}).Count()
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}
	if count > 0 {
		ws.Respond(w, http.StatusConflict, "shopify store already connected")
		return
	}

	url := connect.SH.AuthCodeURL(r)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/nearby", Nearby)
	r.Get("/recent", Recent)
	r.Get("/following", ListFollowing)
	r.Get("/all", ListPlaces)
	r.Post("/autocomplete", AutoComplete)

	r.Route("/manage", func(r chi.Router) {
		r.Get("/", ListManagable)
	})

	r.Route("/:placeID", func(r chi.Router) {
		r.Use(PlaceCtx)
		r.Get("/", GetPlace)
		r.Get("/promos", ListPromo)
		r.Get("/connect/shopify", ConnectShopify)

		r.Post("/follow", FollowPlace)
		r.Delete("/follow", UnfollowPlace)
	})

	return r
}
