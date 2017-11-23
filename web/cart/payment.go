package cart

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/pressly/lg"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/lib/stripe"
	"bitbucket.org/moodie-app/moodie-api/web/api"
)

type cartPayment struct {
	Number string `json:"number"`
	Type   string `json:"type"`
	Expiry string `json:"expiry"`
	Name   string `json:"name"`
	CVC    string `json:"cvc"`
}

type cartPaymentRequest struct {
	Payment        *cartPayment      `json:"payment"`
	BillingAddress *data.CartAddress `json:"billingAddress"`
}

func (c *cartPaymentRequest) Bind(r *http.Request) error {
	if c.Payment == nil {
		return errors.New("no payment specified")
	}
	return nil
}

// TODO: Need to do the stripe token exchange on the frontend
// to be PCI complient. This is critical
func CreatePayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cart := ctx.Value("cart").(*data.Cart)

	var payload cartPaymentRequest
	if err := render.Bind(r, &payload); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	expiryYM := strings.Split(payload.Payment.Expiry, "/")
	cardParam := &stripe.CardParams{
		Name:   payload.Payment.Name,
		Number: payload.Payment.Number,
		Month:  expiryYM[0],
		Year:   expiryYM[1],
		CVC:    payload.Payment.CVC,
	}

	checkout := shopify.Checkout{}
	// update billing address, if specified
	if b := payload.BillingAddress; b != nil {
		checkout.BillingAddress = &shopify.CustomerAddress{
			Address1:  b.Address,
			Address2:  b.AddressOpt,
			City:      b.City,
			Country:   b.Country,
			FirstName: b.FirstName,
			LastName:  b.LastName,
			Province:  b.Province,
			Zip:       b.Zip,
		}
		cart.Etc.BillingAddress = payload.BillingAddress

		cardParam.Address1 = b.Address
		cardParam.Address2 = b.AddressOpt
		cardParam.City = b.City
		cardParam.State = b.Province
		cardParam.Zip = b.Zip
		cardParam.Country = b.Country
	} else {
		if s := cart.Etc.ShippingAddress; s != nil {
			cardParam.Address1 = s.Address
			cardParam.Address2 = s.AddressOpt
			cardParam.City = s.City
			cardParam.State = s.Province
			cardParam.Zip = s.Zip
			cardParam.Country = s.Country
		}
	}

	// 1. exchange user credit card information for stripe token
	for placeID, sh := range cart.Etc.ShopifyData {
		ctx = context.WithValue(ctx, connect.StripeAccountKey, sh.PaymentAccountID)
		stripeToken, err := connect.ST.ExchangeToken(ctx, cardParam)
		if err != nil {
			render.Render(w, r, api.ErrStripeProcess(err))
			return
		}
		lg.Warnf("received stripe token for place(%d) on cart(%d)", placeID, cart.ID)

		creds, err := data.DB.ShopifyCred.FindByPlaceID(placeID)
		if err != nil {
			render.Respond(w, r, err)
			return
		}

		cl := shopify.NewClient(nil, creds.AccessToken)
		cl.BaseURL, _ = url.Parse(creds.ApiURL)

		// 2. Check update the shipping line and billing address if set
		ch := checkout
		if m, ok := cart.Etc.ShippingMethods[placeID]; ok {
			ch.ShippingLine = &shopify.ShippingLine{
				Handle: m.Handle,
			}
		}
		if sh.Discount != nil && !sh.Discount.Applicable {
			ch.DiscountCode = ""
		}
		ch.Token = sh.Token
		cc, _, err := cl.Checkout.Update(ctx, &shopify.CheckoutRequest{&ch})
		if err != nil {
			lg.Alert(errors.Wrapf(err, "failed to pay cart(%d). shopify(%v)", cart.ID, placeID))
			continue
		}

		// 3. send stripe payment token to shopify
		u, _ := uuid.NewUUID()
		payment := &shopify.PaymentRequest{
			Payment: &shopify.Payment{
				Amount:      cc.CheckoutPrice.PaymentDue,
				UniqueToken: u.String(),
				PaymentToken: &shopify.PaymentToken{
					PaymentData: stripeToken.ID,
					Type:        shopify.StripeVaultToken,
				},
				RequestDetails: &shopify.RequestDetail{
					IPAddress: stripeToken.ClientIP,
				},
			},
		}

		p, _, err := cl.Checkout.Payment(ctx, sh.Token, payment)
		if err != nil {
			lg.Alertf("payment fail: cart(%d) place(%d) with err %+v", cart.ID, placeID, err)
			render.Respond(w, r, err)
			// TODO: do we return here?
			return
		}
		// 4. save shopify payment id
		sh.PaymentID = p.ID
		sh.PaymentDue = cc.CheckoutPrice.PaymentDue
		sh.TotalTax = atoi(cc.CheckoutPrice.TotalTax)
		sh.TotalPrice = atoi(cc.CheckoutPrice.TotalPrice)
	}

	// mark checkout as has payed
	cart.Status = data.CartStatusPaymentSuccess
	if err := data.DB.Cart.Save(cart); err != nil {
		lg.Alertf("cart (%d) payment save failed with %+v", cart.ID, err)
	}
	// TODO: create a customer on stripe after the first
	// tokenization so we can send stripe customer id moving forward

	render.Render(w, r, presenter.NewCart(ctx, cart))
}
