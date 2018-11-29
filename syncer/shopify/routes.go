package shopify

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"net/url"

	"bitbucket.org/moodie-app/moodie-api/data"
	lib "bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
	"upper.io/db.v3"
)

func ShopifyStoreWhCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		h := r.Header
		shopDomain := h.Get(lib.WebhookHeaderShopDomain)
		if shopDomain == "" {
			render.Respond(w, r, api.ErrBadID)
			return
		}

		// TODO: Use a tld lib
		parts := strings.Split(shopDomain, ".")
		shopID := parts[0]
		ctx := r.Context()

		place, err := storeGet(shopID)
		if err != nil {
			// TODO: this should warrent some form of retry.
			lg.Warnf("webhooks: place(%s) errored with: %+v", shopID, err)
			render.Status(r, http.StatusOK)
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
		topic := h.Get(lib.WebhookHeaderTopic)
		lg.SetEntryField(ctx, "topic", topic)

		// getting the shopify cred
		cred, err := data.DB.ShopifyCred.FindOne(db.Cond{"place_id": place.ID})
		if err != nil {
			return
		}

		// creating the client
		client := lib.NewClient(nil, cred.AccessToken)
		client.BaseURL, err = url.Parse(cred.ApiURL)
		if err != nil {
			return
		}

		// loadup contexts
		ctx = context.WithValue(ctx, "place", place)
		ctx = context.WithValue(ctx, "sync.place", place)
		ctx = context.WithValue(ctx, "sync.topic", topic)
		ctx = context.WithValue(ctx, "shopify.client", client)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("sync.place").(*data.Place)
	topic := ctx.Value("sync.topic").(string)

	// always return OK
	render.Status(r, http.StatusOK)

	// NOTE: new merchants must have an active billing type.
	// TODO: move webhook registration to after billing is accepted
	billing, _ := data.DB.PlaceBilling.FindByPlaceID(place.ID)
	if billing != nil && billing.Status != data.BillingStatusActive {
		lg.SetEntryField(ctx, "error", errors.New("billing inactive"))
		return
	}

	// TODO: implement other webhooks

	switch lib.Topic(topic) {
	case lib.TopicProductListingsAdd,
		lib.TopicProductListingsUpdate,
		lib.TopicProductListingsRemove:
		if err := ProductListingHandler(r); err != nil {
			lg.Alertf("webhook: %s for place(%s) failed with %v", topic, place.Name, err)
			lg.SetEntryField(ctx, "error", err)
			return
		}
	case lib.TopicCollectionListingsAdd,
		lib.TopicCollectionListingsRemove,
		lib.TopicCollectionListingsUpdate:
		if err := CollectionListingHandler(r); err != nil {
			lg.Alertf("webhook: %s for place(%s) failed with %v", topic, place.Name, err)
			lg.SetEntryField(ctx, "error", err)
			return
		}
	case lib.TopicAppUninstalled, lib.TopicShopUpdate:
		if err := ShopHandler(r); err != nil {
			lg.Alertf("webhook: %s failed with %v", topic, err)
			lg.SetEntryField(ctx, "error", err)
			return
		}
	case lib.TopicCheckoutsUpdate:
		CheckoutHandler(r)
	default:
		lg.Infof("ignoring webhook topic %s for place(id=%d)", topic, place.ID)
	}
}

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(ShopifyStoreWhCtx)
	r.Post("/", WebhookHandler)

	return r
}
