package shopify

import (
	"context"
	"net/http"
	"net/url"

	"github.com/go-chi/render"
	"github.com/pressly/lg"
	db "upper.io/db.v3"

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
	defer r.Body.Close()

	// TODO: implement other webhooks
	ctx = context.WithValue(ctx, "sync.type", shopify.Topic(topic))

	switch shopify.Topic(topic) {
	case shopify.TopicProductListingsAdd:
		ctx = context.WithValue(ctx, "sync.list", []*shopify.ProductList{wrapper.ProductListing})
		ctx = context.WithValue(ctx, "category.cache", categoryCache)
		ctx = context.WithValue(ctx, "category.blacklist", blacklistCache)
		if err := sync.ShopifyProductListingsCreate(ctx); err != nil {
			lg.Warnf("webhook: productAdd for place(%s) failed with %v", place.Name, err)
			return
		}
	case shopify.TopicProductListingsUpdate:
		ctx = context.WithValue(ctx, "sync.list", []*shopify.ProductList{wrapper.ProductListing})
		if err := sync.ShopifyProductListingsUpdate(ctx); err != nil {
			lg.Warnf("webhook: productUpdate for place(%s) failed with %v", place.Name, err)
			return
		}
	case shopify.TopicProductListingsRemove:
		ctx = context.WithValue(ctx, "sync.list", []*shopify.ProductList{wrapper.ProductListing})
		if err := sync.ShopifyProductListingsRemove(ctx); err != nil {
			lg.Warnf("webhook: productRemove for place(%s) failed with %v", place.Name, err)
			return
		}
	case shopify.TopicAppUninstalled:
		lg.Infof("app uninstalled for place(id=%d)", place.ID)
		cred, err := data.DB.ShopifyCred.FindByPlaceID(place.ID)
		if err != nil {
			lg.Warnf("webhook: appUninstall for place(%s) cred failed with %v", place.Name, err)
			return
		}
		api := shopify.NewClient(nil, cred.AccessToken)
		api.BaseURL, _ = url.Parse(cred.ApiURL)

		webhooks, err := data.DB.Webhook.FindByPlaceID(place.ID)
		if err != nil {
			lg.Warnf("webhook: appUninstall for place(%s) webhook failed with %v", place.Name, err)
			return
		}

		for _, wh := range webhooks {
			api.Webhook.Delete(ctx, wh.ExternalID)
			data.DB.Webhook.Delete(wh)
		}

		// remove the credential
		data.DB.ShopifyCred.Delete(cred)

		// set place status to inactive
		place.Status = data.PlaceStatusInActive
		data.DB.Place.Save(place)

		// clean up products
		data.DB.Product.Find(db.Cond{"place_id": place.ID}).Delete()

		lg.Warnf("webhook: place(%s) uninstalled Localyyz", place.Name)
	default:
		lg.Infof("ignoring webhook topic %s for place(id=%d)", topic, place.ID)
	}

	render.Status(r, http.StatusOK)
	return
}
