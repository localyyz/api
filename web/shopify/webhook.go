package shopify

import (
	"context"
	"net/http"
	"net/url"

	"github.com/goware/lg"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
)

func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	h := r.Header
	shopDomain := h.Get("X-Shopify-Shop-Domain")
	u, err := url.Parse(shopDomain)
	if err != nil {
		ws.Respond(w, http.StatusBadRequest, err)
		return
	}

	place, err := data.DB.Place.FindByShopifyID(u.Host)
	if err != nil {
		ws.Respond(w, http.StatusBadRequest, err)
		return
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, "place", place)

	topic := h.Get(shopify.WebhookHeaderTopic)
	switch shopify.Topic(topic) {
	case shopify.TopicProductsCreate:
		var p shopify.Product
		if err := ws.Bind(r.Body, &p); err != nil {
			ws.Respond(w, http.StatusUnprocessableEntity, err)
			return
		}

		product, promos := getProductPromo(ctx, &p)
		if err := data.DB.Product.Save(product); err != nil {
			ws.Respond(w, http.StatusInternalServerError, err)
			return
		}

		for _, v := range promos {
			if err := data.DB.Promo.Save(v); err != nil {
				ws.Respond(w, http.StatusInternalServerError, err)
				return
			}
		}
	case shopify.TopicProductsUpdate:
		var p shopify.Product
		if err := ws.Bind(r.Body, &p); err != nil {
			ws.Respond(w, http.StatusUnprocessableEntity, err)
			return
		}
		// look up by external id
		_, err := data.DB.Product.FindByExternalID(p.Handle)
		if err != nil {
			ws.Respond(w, http.StatusInternalServerError, err)
			return
		}

	default:
		lg.Infof("ignoring webhook topic %s for %s", topic, h.Get(shopify.WebhookHeaderShopDomain))
	}
	return
}
