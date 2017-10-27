package shopify

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/go-chi/render"
	"github.com/pressly/lg"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/lib/sync"
	"bitbucket.org/moodie-app/moodie-api/web/api"
)

type shopifyWebhookRequest struct {
	ProductListing *shopify.ProductList `json:"product_listing"`
}

func (*shopifyWebhookRequest) Bind(r *http.Request) error {
	return nil
}

func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)
	topic := ctx.Value("sync.topic").(string)

	wrapper := new(shopifyWebhookRequest)
	if err := render.Bind(r, wrapper); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	defer func() {
		// log the call for future use
		b, _ := json.Marshal(wrapper)
		lg.Infof(`{"topic":"%s","place_id":%d,"data":%s}`, topic, place.ID, string(b))
	}()

	go func(wrapper *shopifyWebhookRequest) { // return right away
		// TODO: implement other webhooks
		switch shopify.Topic(topic) {
		case shopify.TopicProductListingsAdd:
			ctx = context.WithValue(ctx, "sync.list", []*shopify.ProductList{wrapper.ProductListing})
			if err := sync.ShopifyProductListings(ctx); err != nil {
				render.Respond(w, r, err)
				return
			}
		case shopify.TopicProductListingsUpdate:
			ctx = context.WithValue(ctx, "sync.list", []*shopify.ProductList{wrapper.ProductListing})
			if err := sync.ShopifyProductListings(ctx); err != nil {
				render.Respond(w, r, err)
				return
			}
		case shopify.TopicAppUninstalled:
			lg.Infof("app uninstalled for place(id=%d)", place.ID)
			cred, err := data.DB.ShopifyCred.FindByPlaceID(place.ID)
			if err != nil {
				render.Respond(w, r, err)
				return
			}
			api := shopify.NewClient(nil, cred.AccessToken)
			api.BaseURL, _ = url.Parse(cred.ApiURL)

			webhooks, err := data.DB.Webhook.FindByPlaceID(place.ID)
			if err != nil {
				render.Respond(w, r, err)
				return
			}

			for _, wh := range webhooks {
				api.Webhook.Delete(ctx, wh.ExternalID)
				data.DB.Webhook.Delete(wh)
			}

			// TODO: archive the place?
			// remove the credential
			data.DB.ShopifyCred.Delete(cred)
		default:
			lg.Infof("ignoring webhook topic %s for place(id=%d)", topic, place.ID)
		}
	}(wrapper)

	render.Status(r, http.StatusOK)
	return
}
