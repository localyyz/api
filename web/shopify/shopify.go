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
	"bitbucket.org/moodie-app/moodie-api/web/api"
	db "upper.io/db.v3"
)

func CredCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		place := ctx.Value("place").(*data.Place)

		creds, err := data.DB.ShopifyCred.FindByPlaceID(place.ID)
		if err != nil {
			render.Render(w, r, api.WrapErr(err))
			return
		}

		ctx = context.WithValue(ctx, "creds", creds)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func ClientCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		creds := ctx.Value("creds").(*data.ShopifyCred)
		//authClient := connect.SH.ClientFromCred(r)
		api := shopify.NewClient(nil, creds.AccessToken)

		api.BaseURL, _ = url.Parse(creds.ApiURL)

		ctx = context.WithValue(ctx, "api", api)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func Connect(w http.ResponseWriter, r *http.Request) {
	shopID := strings.ToLower(chi.URLParam(r, "shopID"))
	place, err := data.DB.Place.FindByShopifyID(shopID)
	if err != nil && err != db.ErrNoMoreRows {
		render.Render(w, r, api.WrapErr(err))
		return
	}

	if place != nil {
		count, err := data.DB.ShopifyCred.Find(db.Cond{"place_id": place.ID}).Count()
		if err != nil {
			render.Render(w, r, api.WrapErr(err))
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

func SyncProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	shopApi := ctx.Value("api").(*shopify.Client)
	productList, _, err := shopApi.Product.List(ctx)
	if err != nil {
		render.Render(w, r, api.WrapErr(err))
		return
	}

	// initial sync up
	for _, p := range productList {
		product, promos := getProductPromo(ctx, p)

		if err := data.DB.Product.Save(product); err != nil {
			render.Render(w, r, api.WrapErr(err))
			return
		}

		for _, v := range promos {
			v.ProductID = product.ID
			if err := data.DB.Promo.Save(v); err != nil {
				render.Render(w, r, api.WrapErr(err))
				return
			}
		}
	}
	render.Respond(w, r, productList)
}
