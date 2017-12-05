package shopify

import (
	"context"
	"net/http"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/web/api"

	"github.com/go-chi/render"
	"github.com/pressly/lg"
)

func ShopifyStoreWhCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		h := r.Header
		shopDomain := h.Get(shopify.WebhookHeaderShopDomain)
		if shopDomain == "" {
			render.Respond(w, r, api.ErrBadID)
			return
		}

		// TODO: Use a tld lib
		parts := strings.Split(shopDomain, ".")
		shopID := parts[0]
		ctx := r.Context()

		place, err := data.DB.Place.FindByShopifyID(shopID)
		if err != nil {
			lg.Alertf("webhooks: place(%s) errored with: %+v", shopID, err)
			render.Respond(w, r, err)
			return
		}
		// log the place context
		lg.SetEntryField(ctx, "place_id", place.ID)

		if place.Status != data.PlaceStatusActive {
			// if not active, return and ignore
			render.Status(r, http.StatusOK)
			return
		}

		// TODO: check HMAC
		topic := h.Get(shopify.WebhookHeaderTopic)

		// loadup contexts
		ctx = context.WithValue(ctx, "place", place)
		ctx = context.WithValue(ctx, "sync.place", place)
		ctx = context.WithValue(ctx, "sync.topic", topic)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}
