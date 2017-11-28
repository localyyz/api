package merchant

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
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

		place, err := data.DB.Place.FindOne(
			db.Cond{
				"shopify_id": shopID,
				"status <>":  data.PlaceStatusInActive,
			},
		)
		if err != nil && err != db.ErrNoMoreRows {
			render.Respond(w, r, err)
			return
		}

		if place == nil {
			// redirect to api to trigger oaut
			apiURL := ctx.Value("api.url").(string)
			redirectURL := fmt.Sprintf("%s/connect?shop=%s", apiURL, shopDomain)
			http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
			return
		}

		ctx = context.WithValue(ctx, "place", place)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func ShopifyChargeCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		place := ctx.Value("place").(*data.Place)

		chargeID, err := strconv.ParseInt(r.URL.Query().Get("charge_id"), 10, 64)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		creds, err := data.DB.ShopifyCred.FindByPlaceID(place.ID)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		sh := shopify.NewClient(nil, creds.AccessToken)
		sh.BaseURL, _ = url.Parse(creds.ApiURL)

		// check if place billing id match
		if place.Billing.ID != chargeID {
			lg.Warnf("merchant(%d) expected billing id (%d) but received (%d)", place.ID, place.Billing.ID, chargeID)
			next.ServeHTTP(w, r)
			return
		}

		if _, _, err := sh.Billing.Get(ctx, place.Billing.Billing); err != nil {
			lg.Warnf("merchant(%d) fetch billing failed with: %+v", place.ID, err)
			next.ServeHTTP(w, r)
			return
		}

		// if the status is "accepted", activate it
		if place.Billing.Billing.Status == shopify.BillingStatusAccepted {
			_, _, err = sh.Billing.Activate(ctx, place.Billing.Billing)
			if err != nil {
				lg.Alertf("merchant (%d) failed to activate billing with: %+v", place.ID, err)
				next.ServeHTTP(w, r)
				return
			}
		}

		// save place and billing
		lg.Warnf("received merchant (%d) billing status %s", place.ID, place.Billing.Billing.Status)
		if err := data.DB.Place.Save(place); err != nil {
			render.Respond(w, r, err)
			return
		}

		ctx = context.WithValue(ctx, "place", place)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}
