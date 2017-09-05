package shopify

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/pressly/chi"
	"github.com/pressly/chi/render"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/lib/sync"
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

func SyncProductList(w http.ResponseWriter, r *http.Request) {
	shopID := strings.ToLower(chi.URLParam(r, "shopID"))
	place, err := data.DB.Place.FindByShopifyID(shopID)
	if err != nil && err != db.ErrNoMoreRows {
		render.Respond(w, r, err)
		return
	}

	creds, err := data.DB.ShopifyCred.FindByPlaceID(place.ID)
	if err != nil {
		render.Respond(w, r, err)
		return
	}
	api := shopify.NewClient(nil, creds.AccessToken)
	api.BaseURL, _ = url.Parse(creds.ApiURL)

	ctx := r.Context()
	productList, _, _ := api.ProductList.Get(ctx)
	ctx = context.WithValue(ctx, "sync.list", productList)
	ctx = context.WithValue(ctx, "sync.place", place)
	sync.ShopifyProductListings(ctx)

}
