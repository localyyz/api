package merchant

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
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

		place, err := data.DB.Place.FindByShopifyID(shopID)
		if err != nil && err != db.ErrNoMoreRows {
			render.Respond(w, r, err)
			return
		}

		if place == nil {
			// redirect to api to trigger oaut
			apiURL := ctx.Value("api.url").(string)
			redirectURL := fmt.Sprintf("%s/connect?shop=%s", apiURL, shopDomain)
			lg.Warn(redirectURL)
			http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
			return
		}

		ctx = context.WithValue(ctx, "place", place)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}
