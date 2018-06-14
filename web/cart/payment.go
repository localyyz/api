package cart

import (
	"context"
	"net/http"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/shopper"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/render"
	"github.com/pkg/errors"
	"github.com/pressly/lg"
)

type cartPaymentRequest struct {
	Card           *shopper.PaymentCard `json:"payment"`
	BillingAddress *data.CartAddress    `json:"billingAddress"`
}

func (c *cartPaymentRequest) Bind(r *http.Request) error {
	if c.Card == nil {
		return errors.New("no payment card found")
	}
	c.Card.Name = strings.TrimSpace(c.Card.Name)
	return nil
}

// Start payment process on a shopping cart + checkouts
//
// 'cart' is a collection of 'checkouts' from different stores
func CreatePayments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cart := ctx.Value("cart").(*data.Cart)

	var payload cartPaymentRequest
	if err := render.Bind(r, &payload); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	if cart.Status > data.CartStatusCheckout {
		render.Render(w, r, api.ErrInvalidRequest(ErrInvalidStatus))
		return
	}

	checkouts, err := data.DB.Checkout.FindAllByCartID(cart.ID)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	ctx = context.WithValue(ctx, shopper.BillingAddressCtxKey, cart.BillingAddress)
	ctx = context.WithValue(ctx, shopper.RequestIPCtxKey, r.RemoteAddr)
	ctx = context.WithValue(ctx, shopper.PaymentCardCtxKey, payload.Card)

	var paymentErrors []error
	for _, c := range checkouts {
		p := shopper.NewPayment(ctx, c)
		if err := p.Do(nil); err != nil || p.Err != nil {
			if err != nil {
				lg.Alertf("checkout (%d) payment: %v", c.ID, err)
				// some internal server error, return right away
				render.Respond(w, r, err)
				return
			} else {
				paymentErrors = append(paymentErrors, err)
			}
		}
	}

	if len(paymentErrors) == 0 {
		cart.Status = data.CartStatusComplete
		if err := data.DB.Cart.Save(cart); err != nil {
			lg.Alertf("failed to save cart status, cart id: %d", cart.ID)
		}
	}

	// TODO: return all errors
	presented := presenter.NewCart(ctx, cart)
	for _, lastError := range paymentErrors {
		presented.HasError = true
		presented.Error = lastError.Error()

		render.Status(r, http.StatusBadRequest)
		break
	}

	// give us a nice alert
	if !presented.HasError && cart.Status == data.CartStatusComplete {
		lg.Alertf("%s just completed a purchase of $%d! hoorah!", presented.ShippingAddress.FirstName, presented.TotalPrice/100.0)
	}

	render.Respond(w, r, presented)
}
