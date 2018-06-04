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
		return errors.New("payment method not specified specified")
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
		if err := p.Do(nil); err != nil {
			lg.Alertf("checkout(%d) payment: %v", c.ID, err)
			paymentErrors = append(paymentErrors, err)
		}
	}
	// TODO: something needs to happen here if there's error
	if len(paymentErrors) == 0 {
		cart.Status = data.CartStatusComplete
	}

	render.Respond(w, r, presenter.NewCart(ctx, cart))
}
