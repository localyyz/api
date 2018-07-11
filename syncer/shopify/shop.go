package shopify

import (
	"net/http"
	"net/url"

	"bitbucket.org/moodie-app/moodie-api/data"
	lib "bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
)

type shopWrapper struct {
	*lib.Shop
}

func (w *shopWrapper) Bind(r *http.Request) error {
	return nil
}

func ShopHandler(r *http.Request) error {
	wrapper := new(shopWrapper)
	if err := render.Bind(r, wrapper); err != nil {
		return api.ErrInvalidRequest(err)
	}
	defer r.Body.Close()

	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)

	switch lib.Topic(ctx.Value("sync.topic").(string)) {
	case lib.TopicAppUninstalled:
		lg.Infof("app uninstalled for place(id=%d)", place.ID)
		cred, err := data.DB.ShopifyCred.FindByPlaceID(place.ID)
		if err != nil {
			lg.Warnf("webhook: appUninstall for place(%s) cred failed with %v", place.Name, err)
			return err
		}
		api := lib.NewClient(nil, cred.AccessToken)
		api.BaseURL, _ = url.Parse(cred.ApiURL)

		webhooks, err := data.DB.Webhook.FindByPlaceID(place.ID)
		if err != nil {
			lg.Warnf("webhook: appUninstall for place(%s) webhook failed with %v", place.Name, err)
			return err
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

		lg.Alertf("webhook: place(%s) uninstalled Localyyz", place.Name)
	case lib.TopicShopUpdate:
		if place.Plan != wrapper.PlanName {
			lg.Alertf("webhook: place(%s) is now %s", place.Name, wrapper.PlanName)
			place.Plan = wrapper.PlanName
			if place.Plan == "dormant" {
				place.Status = data.PlaceStatusInActive
				// TODO: Should we clean up everything? probably
			}
		}

		place.Name = wrapper.Name
		place.Phone = wrapper.Phone
		place.Currency = wrapper.Currency

		data.DB.Place.Save(place)
	}
	return nil
}
