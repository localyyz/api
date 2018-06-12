package cart

import (
	"context"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/shopper"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
)

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

// Start checkout process on a shopping cart
//
// 'cart' is a collection of 'checkouts' from different stores
func CreateCheckouts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cart := ctx.Value("cart").(*data.Cart)
	user := ctx.Value("session.user").(*data.User)

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
	ctx = context.WithValue(ctx, shopper.EmailCtxKey, user.Email)
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
				lg.Alertf("checkout(%d): %v %v", c.ID, err)
				// some internal server error, return right away
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

			switch e.ErrCode {
			case shopper.CheckoutErrorCodeNoShipping, shopper.CheckoutErrorCodeShippingAddress:
				presented.ShippingAddress.HasError = true
				presented.ShippingAddress.Error = e.Err
			case shopper.CheckoutErrorCodeBillingAddress:
				presented.BillingAddress.HasError = true
				presented.BillingAddress.Error = e.Err
			}

			for _, ci := range presented.CartItems {
				ci.HasError = ci.Variant.OfferID == e.ItemID
				ci.Error = e.Err
			}
			render.Status(r, http.StatusBadRequest)
		}
		break
	}
	render.Render(w, r, presented)
}
