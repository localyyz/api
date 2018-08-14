package cart

import (
	"context"
	"net/http"
	"strconv"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/shopper"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
)

func CheckoutCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		cart := ctx.Value("cart").(*data.Cart)

		checkoutID, err := strconv.ParseInt(chi.URLParam(r, "checkoutID"), 10, 64)
		if err != nil {
			render.Render(w, r, api.ErrBadID)
			return
		}

		checkout, err := data.DB.Checkout.FindOne(
			db.Cond{
				"id":      checkoutID,
				"cart_id": cart.ID,
			},
		)
		if err != nil {
			render.Respond(w, r, err)
			return
		}

		ctx = context.WithValue(ctx, "checkout", checkout)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func validateCart(ctx context.Context) error {
	cart := ctx.Value("cart").(*data.Cart)

	if cart.Email == "" {
		return ErrInvalidEmail
	}

	if cart.ShippingAddress == nil {
		return ErrInvalidShipping
	}

	if cart.BillingAddress == nil {
		return ErrInvalidBilling
	}
	return nil
}

type checkoutRequest struct {
	Discount string `json:"discount,omitempty"`
}

func (c *checkoutRequest) Bind(r *http.Request) error {
	return nil
}

// Start checkout process on a shopping cart
//
// 'cart' is a collection of 'checkouts' from different stores
func UpdateCheckout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	checkout := ctx.Value("checkout").(*data.Checkout)

	var payload checkoutRequest
	if err := render.Bind(r, &payload); err != nil {
		render.Respond(w, r, err)
		return
	}

	if len(payload.Discount) != 0 {
		checkout.DiscountCode = payload.Discount
		if err := data.DB.Checkout.Save(checkout); err != nil {
			render.Respond(w, r, err)
			return
		}
	}

	render.Respond(w, r, presenter.NewCheckout(ctx, checkout))
}

// Start checkout process on a shopping cart
//
// 'cart' is a collection of 'checkouts' from different stores
func CreateCheckouts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cart := ctx.Value("cart").(*data.Cart)

	if cart.Status > data.CartStatusCheckout {
		render.Render(w, r, api.ErrInvalidRequest(ErrInvalidStatus))
		return
	}

	// TODO verify cart checkout fields
	if err := validateCart(ctx); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	// add to ctx
	ctx = context.WithValue(ctx, shopper.EmailCtxKey, cart.Email)
	ctx = context.WithValue(ctx, shopper.ShippingAddressCtxKey, cart.ShippingAddress)
	ctx = context.WithValue(ctx, shopper.BillingAddressCtxKey, cart.BillingAddress)

	checkouts, err := data.DB.Checkout.FindAllByCartID(cart.ID)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	var checkoutErrors []error
	for _, c := range checkouts {
		req := shopper.NewCheckout(ctx, c)
		if err := req.Do(nil); err != nil || req.Err != nil {
			if err != nil {
				// some internal server error, return right away
				lg.Alertf("[internal] checkout(%d): %v", c.ID, err)
				render.Respond(w, r, err)
				return
			} else {
				checkoutErrors = append(checkoutErrors, req.Err)
			}
		}
	}

	presented := presenter.NewCart(ctx, cart)
	// TODO: return all errors
	for _, lastError := range checkoutErrors {
		if e, ok := lastError.(*shopper.CheckoutError); e != nil && ok {
			presented.HasError = true
			presented.Error = e.Err.Error()
			presented.ErrorCode = uint32(e.ErrCode)

			lg.Warnf("cart(%d) err: %s", presented.ID, presented.Error)

			switch e.ErrCode {
			case shopper.CheckoutErrorCodeNoShipping, shopper.CheckoutErrorCodeShippingAddress:
				presented.ShippingAddress.HasError = true
				presented.ShippingAddress.Error = e.Err
			case shopper.CheckoutErrorCodeBillingAddress:
				presented.BillingAddress.HasError = true
				presented.BillingAddress.Error = e.Err
			}
			if itemID := e.ItemID; itemID != 0 {
				for _, ci := range presented.CartItems {
					if ci.Variant.OfferID == itemID {
						ci.HasError = true
						ci.Err = e.Err.Error()
					}
				}
			}
			render.Status(r, http.StatusBadRequest)
		}
		break
	}
	render.Render(w, r, presented)
}
