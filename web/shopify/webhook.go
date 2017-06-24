package shopify

import (
	"context"
	"net/http"
	"net/url"

	"github.com/goware/lg"
	"github.com/pressly/chi/render"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/lib/sync"
	"bitbucket.org/moodie-app/moodie-api/web/api"
)

type shopifyWebhookRequest struct {
	*shopify.Product
}

func (*shopifyWebhookRequest) Bind(r *http.Request) error {
	return nil
}

func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	h := r.Header
	shopDomain := h.Get("X-Shopify-Shop-Domain")
	u, err := url.Parse(shopDomain)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	go func() { // return right away
		place, err := data.DB.Place.FindByShopifyID(u.Host)
		if err != nil {
			render.Respond(w, r, err)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, "sync.place", place)

		topic := h.Get(shopify.WebhookHeaderTopic)
		switch shopify.Topic(topic) {
		case shopify.TopicProductsCreate:
			p := &shopifyWebhookRequest{}
			if err := render.Bind(r, p); err != nil {
				render.Render(w, r, api.ErrInvalidRequest(err))
				return
			}

			ctx = context.WithValue(ctx, "sync.list", []*shopify.Product{p.Product})
			if err := sync.ShopifyProducts(ctx); err != nil {
				render.Respond(w, r, err)
				return
			}
		case shopify.TopicProductsUpdate:
			p := &shopifyWebhookRequest{}
			if err := render.Bind(r, p); err != nil {
				render.Respond(w, r, err)
				return
			}
			// look up by external id
			_, err := data.DB.Product.FindByExternalID(p.Handle)
			if err != nil {
				render.Respond(w, r, err)
				return
			}
		default:
			lg.Infof("ignoring webhook topic %s for %s", topic, h.Get(shopify.WebhookHeaderShopDomain))
		}

	}()

	return
}
