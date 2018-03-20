package shopify

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/pressly/lg"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/web/shopify/webhook"
)

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
		return
	}

	// TODO: implement other webhooks

	switch shopify.Topic(topic) {
	case shopify.TopicProductListingsAdd,
		shopify.TopicProductListingsUpdate,
		shopify.TopicProductListingsRemove:
		if err := webhooks.ProductListingHandler(r); err != nil {
			lg.Warnf("webhook: %s for place(%s) failed with %v", topic, place.Name, err)
			return
		}
	case shopify.TopicAppUninstalled, shopify.TopicShopUpdate:
		webhooks.ShopHandler(r)
	case shopify.TopicCheckoutsUpdate:
		webhooks.CheckoutHandler(r)
	default:
		lg.Infof("ignoring webhook topic %s for place(id=%d)", topic, place.ID)
	}

}
