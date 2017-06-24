package shopify

import (
	"context"
	"net/http"
	"strings"

	"github.com/pressly/chi"
	"github.com/pressly/chi/render"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	db "upper.io/db.v3"
)

func Connect(w http.ResponseWriter, r *http.Request) {
	shopID := strings.ToLower(chi.URLParam(r, "shopID"))
	place, err := data.DB.Place.FindByShopifyID(shopID)
	if err != nil && err != db.ErrNoMoreRows {
		render.Respond(w, r, err)
		return
	}

	if place != nil {
		count, err := data.DB.ShopifyCred.Find(db.Cond{"place_id": place.ID}).Count()
		if err != nil {
			render.Respond(w, r, err)
			return
		}
		if count > 0 {
			render.Render(w, r, api.ErrConflictStore)
			return
		}
	} else {
		place = &data.Place{ShopifyID: shopID}
	}

	ctx := context.WithValue(r.Context(), "place", place)
	url := connect.SH.AuthCodeURL(ctx)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
