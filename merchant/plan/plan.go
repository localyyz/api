package plan

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

func Routes() chi.Router {
	r := chi.NewRouter()

	r.With(RecurringChargeCtx).Get("/cb", Callback)
	r.With(BillingPlanTypeCtx).Post("/", CreatePlan)

	return r
}

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
		// continue on normally if not found
		billing, err := data.DB.PlaceBilling.FindOne(
			db.Cond{
				"place_id":    place.ID,
				"external_id": chargeID,
			},
		)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		ctx = context.WithValue(ctx, "place.billing", billing)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

// handle the accepted/declined billing callback
func handleBillingCallback(ctx context.Context, place *data.Place) error {
	billing, ok := ctx.Value("place.billing").(*data.PlaceBilling)
	if !ok {
		return api.ErrInvalidRequest(nil)
	}

	cl := ctx.Value("shopify.client").(*shopify.Client)
	// fetch shopify billing and check if the status is "accpted"
	shopifyBilling := &shopify.Billing{
		ID:   billing.ExternalID,
		Type: shopify.BillingTypeRecurring,
	}
	if _, _, err := cl.Billing.Get(ctx, shopifyBilling); err != nil {
		return errors.Wrap(err, "failed to fetch billing")
	}

	// billing was accepted
	if shopifyBilling.Status == shopify.BillingStatusAccepted {
		// activate accepted billing
		if _, _, err := cl.Billing.Activate(ctx, shopifyBilling); err != nil {
			return errors.Wrap(err, "failed to activate billing")
		}
	}

	// save billing status (could be active/declined)
	billing.Status = data.BillingStatus(shopifyBilling.Status)

	lg.Alertf("merchant(%d) billing status is %v", place.ID, billing.Status)
	// save place and billing
	ctx = context.WithValue(ctx, "place.billing", billing)
	if err := data.DB.PlaceBilling.Save(billing); err != nil {
		return errors.Wrap(err, "failed to save place billing")
	}
	return nil
}

func finalize(ctx context.Context) error {
	billing := ctx.Value("place.billing").(*data.PlaceBilling)
	if billing.Status != data.BillingStatusActive {
		// nothing to do. return
		return nil
	}

	plan, err := data.DB.BillingPlan.FindByID(billing.PlanID)
	if err != nil {
		return errors.Wrap(err, "fetch billing plan")
	}

	// finalize initial subscription charge
	c, err := data.DB.PlaceCharge.FindActiveSubscriptionCharge(billing.PlaceID)
	if err != nil && err != db.ErrNoMoreRows {
		return errors.Wrap(err, "fetch subscription charge")
	}
	// already charged, return
	if c != nil {
		return nil
	}

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
	expireAt := time.Now().Add(Year)
	dbCharge.ExpireAt = &expireAt

	// save charge to the database
	// TODO: safe guards? should have retries etc
	err = data.DB.PlaceCharge.Save(dbCharge)
	if err != nil {
		return errors.Wrap(err, "save subscription charge")
	}

	return nil
}

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

	if err := handleBillingCallback(ctx, place); err != nil {
		lg.Alertf("merchant(%d) billing callback failed: %v", place.ID, err)
	}

	if err := finalize(ctx); err != nil {
		lg.Alertf("merchant(%d) billing callback failed: %v", place.ID, err)
	}

	appName := ctx.Value("shopify.appname").(string)
	http.Redirect(w, r, fmt.Sprintf("https://%s.myshopify.com/admin/apps/%s", place.ShopifyID, appName), http.StatusTemporaryRedirect)
}

func CreatePlan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	planType := ctx.Value("billing.type").(data.BillingPlanType)
	place := ctx.Value("place").(*data.Place)

	if !place.PlanEnabled {
		return
	}

	defaultPlan, err := data.DB.BillingPlan.FindDefaultByType(planType)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	// create a recurring billing for transaction and commission charges
	billing := &shopify.Billing{
		Type:         shopify.BillingTypeRecurring,
		Name:         defaultPlan.Name,
		Terms:        defaultPlan.Terms,
		ReturnUrl:    fmt.Sprintf("https://merchant.localyyz.com/plan/cb"),
		Price:        "0",
		CappedAmount: "10000",
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
