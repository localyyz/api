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

	go func(wrapper *shopifyWebhookRequest) { // return right away
		// TODO: implement other webhooks
		switch shopify.Topic(topic) {
		case shopify.TopicProductListingsAdd:
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

			// remove all webhooks
			//webhooks, _, _ := api.Webhook.List(ctx)

		default:
			lg.Infof("ignoring webhook topic %s for place(id=%d)", topic, place.ID)
		}
	}(wrapper)

	return
}
