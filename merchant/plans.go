package merchant

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"bitbucket.org/moodie-app/moodie-api/data"
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

	cl := ctx.Value("shopify.client").(*shopify.Client)
	// fetch shopify billing and check if the status is "accepted"
	shopifyBilling := &shopify.Billing{
		ID:   billing.ExternalID,
		Type: shopify.BillingTypeRecurring,
	}
	if _, _, err := cl.Billing.Get(ctx, shopifyBilling); err != nil {
		return errors.Wrap(err, "failed to fetch billing")
	}

	// billing was accepted. activate it (required by shopify)
	if shopifyBilling.Status == shopify.BillingStatusAccepted {
		if _, _, err := cl.Billing.Activate(ctx, shopifyBilling); err != nil {
			return errors.Wrap(err, "failed to activate billing")
		}
	}

	// save billing status (could be active/declined)
	billing.Status = data.BillingStatus(shopifyBilling.Status)

	lg.Alertf("merchant(%d) billing status is %v", billing.PlaceID, billing.Status)
	if err := data.DB.PlaceBilling.Save(billing); err != nil {
		return errors.Wrap(err, "failed to save place billing")
	}
	return nil
}

// finalize finishes up the billing callback.
// for example:
//  - submit a "usage" charge for plans not billed monthly (TODO/NOTE: shopify
//  will take care of this in the future)
//  - TODO: install webhooks and start accepting products
func finalize(ctx context.Context) error {
	plan := ctx.Value("place.plan").(*data.BillingPlan)
	billing := ctx.Value("place.billing").(*data.PlaceBilling)
	if billing.Status != data.BillingStatusActive {
		// nothing to do. return
		return nil
	}

	// if the billing type is Annual / (TODO: quaterly), submit
	// an initial usage charge as the subscription fee.
	// NOTE/TODO: shopify is planning to natively support other than monthly
	// subscription types, so this can be handled on their end
	if plan.RecurringPrice == 0.0 {
		// return early because we did not match above requirement
		return nil
	}
	if plan.BillingType != data.BillingTypeAnnual {
		// return early because we did not match above requirement
		return nil
	}

	// check if we have any active subscription charge already
	// if we do, return
	c, err := data.DB.PlaceCharge.FindActiveSubscriptionCharge(billing.PlaceID)
	if err != nil && err != db.ErrNoMoreRows {
		// something went wrong, return right away
		return errors.Wrap(err, "fetch subscription charge")
	}
	// already charged, return
	if c != nil {
		return nil
	}

	// create the shopify usage charge.
	// TODO: support quaterly subscription types
	shopifyCharge := &shopify.UsageCharge{
		RecurringApplicationChargeID: billing.ExternalID,
		Description:                  plan.Name,
		Price:                        fmt.Sprintf("%.2f", plan.RecurringPrice),
	}
	cl := ctx.Value("shopify.client").(*shopify.Client)
	if _, _, err := cl.Billing.CreateUsageCharge(ctx, shopifyCharge); err != nil {
		return errors.Wrap(err, "creating subscription charge")
	}
	dbCharge := &data.PlaceCharge{
		PlaceID:    billing.PlaceID,
		ExternalID: shopifyCharge.ID,
		Amount:     plan.RecurringPrice,
		ChargeType: data.ChargeTypeSubscription,
	}

	// calculate the end date
	// hard coded to "Year", because of annual
	expireAt := time.Now().Add(Year)
	dbCharge.ExpireAt = &expireAt

	// save charge to the database
	// TODO: safe guards? should have retries etc
	err = data.DB.PlaceCharge.Save(dbCharge)
	if err != nil {
		return errors.Wrap(err, "save subscription charge")
	}

	lg.Alertf("merchant(%d) finalized billing type (%s): %s", billing.PlaceID, plan.PlanType, billing.Status)
	return nil
}

// Shopify routes to this endpoint after user accepts/rejects the presented
// recurring charge.
//
// Handling is split into two parts:
//
// - handle billing callback
// - finalize
func Callback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)

	creds, err := data.DB.ShopifyCred.FindByPlaceID(place.ID)
	if err != nil {
		lg.Alertf("merchant(%d) billing callback failed: %v", place.ID, err)
		return
	}
	cl := shopify.NewClient(nil, creds.AccessToken)
	cl.BaseURL, _ = url.Parse(creds.ApiURL)
	cl.Debug = true
	ctx = context.WithValue(ctx, "shopify.client", cl)

	if err := activateBilling(ctx); err != nil {
		lg.Alertf("merchant(%d) billing activate failed: %v", place.ID, err)
	}

	if err := finalize(ctx); err != nil {
		lg.Alertf("merchant(%d) billing finalize failed: %v", place.ID, err)
	}

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

	creds, err := data.DB.ShopifyCred.FindByPlaceID(place.ID)
	if err != nil {
		render.Respond(w, r, err)
		return
	}
	client := shopify.NewClient(nil, creds.AccessToken)
	client.BaseURL, _ = url.Parse(creds.ApiURL)

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
