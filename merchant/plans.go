package merchant

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/pkg/errors"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
)

var (
	Year = 365 * 24 * time.Hour
)

func BillingPlanTypeCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var billing *data.BillingPlan
		if err := json.NewDecoder(r.Body).Decode(&billing); err != nil {
			render.Render(w, r, api.ErrInvalidRequest(err))
			return
		}

		ctx = context.WithValue(ctx, "billing.type", billing.PlanType)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func RecurringChargeCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		place := ctx.Value("place").(*data.Place)

		// recurring charge id
		chargeID, err := strconv.ParseInt(r.URL.Query().Get("charge_id"), 10, 64)
		if err != nil {
			render.Respond(w, r, api.ErrInvalidRequest(err))
			return
		}

		// fetch the place with the charge id (external_id)
		billing, err := data.DB.PlaceBilling.FindOne(
			db.Cond{
				"place_id":    place.ID,
				"external_id": chargeID,
			},
		)
		if err != nil {
			render.Render(w, r, api.ErrInvalidChargeID)
			return
		}

		plan, err := data.DB.BillingPlan.FindByID(billing.PlanID)
		if err != nil {
			render.Render(w, r, api.ErrInvalidBillingPlan)
			return
		}

		ctx = context.WithValue(ctx, "place.billing", billing)
		ctx = context.WithValue(ctx, "place.plan", plan)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func planRoutes() chi.Router {
	r := chi.NewRouter()

	r.With(RecurringChargeCtx).Get("/cb", Callback)
	r.With(BillingPlanTypeCtx).Post("/", CreatePlan)

	return r
}

// handle the accepted/declined billing callback
func activateBilling(ctx context.Context) error {
	billing := ctx.Value("place.billing").(*data.PlaceBilling)
	place := ctx.Value("place").(*data.Place)
	client, err := connect.GetShopifyClient(billing.PlaceID)
	if err != nil {
		return err
	}

	// fetch shopify billing and check if the status is "accepted"
	shopifyBilling := &shopify.Billing{
		ID:   billing.ExternalID,
		Type: shopify.BillingTypeRecurring,
	}
	if _, _, err := client.Billing.Get(ctx, shopifyBilling); err != nil {
		return errors.Wrap(err, "failed to fetch billing")
	}

	// billing was accepted. activate the plan and merchant status
	if shopifyBilling.Status == shopify.BillingStatusAccepted {
		// activate the recurring billing plan
		if _, _, err := client.Billing.Activate(ctx, shopifyBilling); err != nil {
			return errors.Wrap(err, "failed to activate billing")
		}
		// finalize place status to active.
		place.Status = data.PlaceStatusActive
		if err := data.DB.Place.Save(place); err != nil {
			return err
		}
	}

	// save billing status (could be active/declined)
	billing.Status = data.BillingStatus(shopifyBilling.Status)
	if err := data.DB.PlaceBilling.Save(billing); err != nil {
		return errors.Wrap(err, "failed to save place billing")
	}
	connect.SL.Notify(
		"store",
		fmt.Sprintf(
			"merchant(%d: %s) billing plan (%d) is now %v",
			place.ID,
			place.Name,
			billing.PlanID,
			billing.Status,
		),
	)
	// post merchant create to zapier for syncing to google sheets the new
	// billing status
	connect.ZP.Post("merchant-create", presenter.NewPlaceApproval(place))
	return nil
}

// Shopify routes to this endpoint after user accepts/rejects the presented
// recurring charge.
func Callback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)

	if err := activateBilling(ctx); err != nil {
		lg.Alertf("merchant(%d) billing activate failed: %v", place.ID, err)
	}

	// redirect back to dashboard
	appName := ctx.Value("shopify.appname").(string)
	http.Redirect(w, r, fmt.Sprintf("https://%s.myshopify.com/admin/apps/%s", place.ShopifyID, appName), http.StatusTemporaryRedirect)
}

// Handles post request when user clicks on one of the billing types
func CreatePlan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	planType := ctx.Value("billing.type").(data.BillingPlanType)
	place, ok := ctx.Value("place").(*data.Place)

	if !ok || !place.PlanEnabled {
		return
	}

	defaultPlan, err := data.DB.BillingPlan.FindDefaultByType(planType)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	// create a recurring billing for transaction and commission charges
	prc := fmt.Sprintf("%.2f", defaultPlan.RecurringPrice)
	billing := &shopify.Billing{
		Type:         shopify.BillingTypeRecurring,
		Name:         defaultPlan.Name,
		Terms:        defaultPlan.Terms,
		ReturnUrl:    "https://merchant.localyyz.com/plan/cb",
		Price:        prc,
		CappedAmount: prc,
	}
	if debug := ctx.Value("debug").(bool); debug {
		billing.Test = true
		billing.ReturnUrl = fmt.Sprintf("https://%s/plan/cb", r.Host)
	}

	client, err := connect.GetShopifyClient(place.ID)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	if _, _, err := client.Billing.Create(ctx, billing); err != nil {
		render.Respond(w, r, errors.Wrapf(err, "failed to create billing for %s", place.ShopifyID))
		return
	}

	placeBilling := &data.PlaceBilling{
		PlaceID:    place.ID,
		PlanID:     defaultPlan.ID,
		Status:     data.BillingStatusPending,
		ExternalID: billing.ID,
	}
	if err := data.DB.PlaceBilling.Save(placeBilling); err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Respond(w, r, billing)
}
