package merchant

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"github.com/go-chi/render"
	db "upper.io/db.v3"
)

func ShopifyShopCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		q := r.URL.Query()
		shopDomain := q.Get("shop")

		if len(shopDomain) == 0 {
			render.Status(r, http.StatusNotFound)
			render.Respond(w, r, "")
			return
		}

		// TODO: Use a tld lib
		parts := strings.Split(shopDomain, ".")
		shopID := parts[0]

		place, err := data.DB.Place.FindOne(
			db.Cond{
				"shopify_id": shopID,
			},
		)
		if err != nil && err != db.ErrNoMoreRows {
			render.Respond(w, r, err)
			return
		}

		if place == nil {
			// redirect to api to trigger oauth
			apiURL := ctx.Value("api.url").(string)
			redirectURL := fmt.Sprintf("%s/connect?shop=%s", apiURL, shopDomain)
			http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
			return
		}

		creds, err := data.DB.ShopifyCred.FindByPlaceID(place.ID)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		client := shopify.NewClient(nil, creds.AccessToken)
		client.BaseURL, _ = url.Parse(creds.ApiURL)

		ctx = context.WithValue(ctx, "shopify.client", client)
		ctx = context.WithValue(ctx, "place", place)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}
