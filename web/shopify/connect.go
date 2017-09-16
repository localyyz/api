package shopify

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/pressly/chi"
	"github.com/pressly/chi/render"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
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
		count, err := data.DB.ShopifyCred.Find(
			db.Cond{
				"place_id": place.ID,
				"status":   data.ShopifyCredStatusActive,
			},
		).Count()
		if err != nil {
			render.Respond(w, r, err)
			return
		}
		if count > 0 {
			// already connected, redirect the user to shopify admin
			adminUrl := fmt.Sprintf("https://%s.myshopify.com/admin/apps/localyyz", shopID)
			http.Redirect(w, r, adminUrl, http.StatusTemporaryRedirect)
			return
		}
	} else {
		place = &data.Place{ShopifyID: shopID}
	}

	ctx := context.WithValue(r.Context(), "place", place)
	url := connect.SH.AuthCodeURL(ctx)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
