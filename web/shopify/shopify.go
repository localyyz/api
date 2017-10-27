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

		place, err := data.DB.Place.FindByShopifyID(shopID)
		if err != nil {
			lg.Alertf("webhooks: place with domain %s is not found", shopDomain)
			render.Respond(w, r, err)
			return
		}

		// TODO: check HMAC
		topic := h.Get(shopify.WebhookHeaderTopic)

		// loadup contexts
		ctx := r.Context()
		ctx = context.WithValue(ctx, "place", place)
		ctx = context.WithValue(ctx, "sync.place", place)
		ctx = context.WithValue(ctx, "sync.topic", topic)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}
