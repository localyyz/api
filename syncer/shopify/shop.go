package shopify

import (
	"net/http"
	"net/url"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	lib "bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/render"
	"github.com/pkg/errors"
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

		// set place status to "uninstalled"
		place.Status = data.PlaceStatusUninstalled
		if err := data.DB.Place.Save(place); err != nil {
			return errors.Wrapf(err, "webhook: appUninstall place(%s,%d)", place.Name, place.ID)
		}

		// clean up products
		_, err = data.DB.Update("products").
			Set(
				db.Raw("status = ?", data.ProductStatusDeleted),
				db.Raw("deleted_at = NOW()"),
			).
			Where(db.Cond{
				"place_id": place.ID,
			}).
			Exec()
		if err != nil {
			return errors.Wrapf(err, "uninstall place(%d)", place.ID)
		}

		// post merchant create to zapier for syncing to google sheets
		connect.ZP.Post("merchant-create", presenter.NewPlaceApproval(place))
		lg.Alertf("webhook: place(%s) uninstalled Localyyz", place.Name)
	case lib.TopicShopUpdate:
		if place.Status == data.PlaceStatusRejected {
			// NOTE: place was rejected. ignore updates
			return nil
		}
		if place.Plan != wrapper.PlanName {
			lg.Alertf("webhook: place(%s) is now %s", place.Name, wrapper.PlanName)
			place.Plan = wrapper.PlanName

			// update the shop status according to the plan name
			switch wrapper.PlanName {
			case "dormant", "cancelled", "frozen",
				"fraudulent", "enterprise", "starter":
				place.Status = data.PlaceStatusInActive

				// clean up the products. mark as pending
				_, err := data.DB.Update("products").
					Set(db.Raw("status = ?", data.ProductStatusPending)).
					Where(db.Cond{
						"place_id": place.ID,
					}).
					Exec()
				if err != nil {
					return errors.Wrapf(err, "update place(%s,%d)", place.Name, place.ID)
				}
			case "affiliate", "staff", "professional",
				"custom", "shopify_plus", "unlimited",
				"basic", "staff_business", "trial",
				"npo_lite", "npo_full", "business":
				place.Status = data.PlaceStatusActive
			default:
				place.Status = data.PlaceStatusReviewing
			}
		}

		place.Name = wrapper.Name
		place.Phone = wrapper.Phone
		place.Currency = wrapper.Currency

		if err := data.DB.Place.Save(place); err != nil {
			return errors.Wrapf(err, "update place(%s,%d)", place.Name, place.ID)
		}
		// post merchant create to zapier for syncing to google sheets
		connect.ZP.Post("merchant-create", presenter.NewPlaceApproval(place))
	}
	return nil
}
