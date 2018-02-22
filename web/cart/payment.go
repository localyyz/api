package cart

import (
	"context"
	"fmt"
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
	"bitbucket.org/moodie-app/moodie-api/lib/shopify/cardvault"
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
	c.Payment.Name = strings.TrimSpace(c.Payment.Name)
	return nil
}

func (c *cartPaymentRequest) parseStripeParams() *stripe.CardParams {
	expiryYM := strings.Split(c.Payment.Expiry, "/")
	cardParam := &stripe.CardParams{
		Name:   c.Payment.Name,
		Number: c.Payment.Number,
		Month:  expiryYM[0],
		Year:   expiryYM[1],
		CVC:    c.Payment.CVC,
	}
	if b := c.BillingAddress; b != nil {
		cardParam.Address1 = b.Address
		cardParam.Address2 = b.AddressOpt
		cardParam.City = b.City
		cardParam.State = b.Province
		cardParam.Zip = b.Zip
		cardParam.Country = b.Country
	}
	return cardParam
}

func (c *cartPaymentRequest) parseCardVault() *cardvault.CreditCard {
	expiryYM := strings.Split(c.Payment.Expiry, "/")
	nameParts := strings.Split(c.Payment.Name, " ")
	firstName := strings.Join(nameParts[0:len(nameParts)-1], " ")
	lastName := nameParts[len(nameParts)-1]
	return &cardvault.CreditCard{
		FirstName:         firstName,
		LastName:          lastName,
		Number:            c.Payment.Number,
		Month:             expiryYM[0],
		Year:              expiryYM[1],
		VerificationValue: c.Payment.CVC,
	}
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
	}

	// 1. exchange user credit card information for stripe token
	for placeID, sh := range cart.Etc.ShopifyData {
		lg.Infof("payment processing cart(%d) for place(%d)", cart.ID, placeID)
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

		var payment *shopify.PaymentRequest
		// 3. Check if using shopify payments -> use stripe
		if len(sh.ShopifyPaymentAccountID) != 0 {
			// 3.1 Use stripe token
			lg.Infof("sending stripe token for place(%d) on cart(%d)", placeID, cart.ID)
			ctx = context.WithValue(ctx, connect.StripeAccountKey, sh.ShopifyPaymentAccountID)
			stripeToken, err := connect.ST.ExchangeToken(ctx, payload.parseStripeParams())
			if err != nil {
				render.Render(w, r, api.ErrStripeProcess(err))
				return
			}
			lg.Infof("received stripe token for place(%d) on cart(%d)", placeID, cart.ID)
			u, _ := uuid.NewUUID()
			payment = &shopify.PaymentRequest{
				Payment: &shopify.Payment{
					Amount:      cc.PaymentDue,
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
		} else {
			// 3.2 Use shopify vault system
			lg.Infof("sending cartvault token for place(%d) on cart(%d)", placeID, cart.ID)

			u, _ := uuid.NewUUID()
			vaultRequest := &cardvault.PaymentRequest{
				Payment: &cardvault.Payment{
					Amount:      cc.PaymentDue,
					CreditCard:  payload.parseCardVault(),
					UniqueToken: u.String(),
				},
			}
			cardVaultID, _, err := cardvault.AddCard(ctx, vaultRequest)
			if err != nil {
				render.Render(w, r, api.ErrCardVaultProcess(err))
				return
			}
			lg.Infof("received cardvault token %s for place(%d) on cart(%d)", cardVaultID, placeID, cart.ID)
			payment = &shopify.PaymentRequest{
				Payment: &shopify.Payment{
					Amount:      cc.PaymentDue,
					UniqueToken: u.String(),
					SessionID:   cardVaultID,
					RequestDetails: &shopify.RequestDetail{
						IPAddress: r.RemoteAddr,
					},
				},
			}
		}

		// 4. send payment to shopify
		p, _, err := cl.Checkout.Payment(ctx, sh.Token, payment)
		if err != nil {
			lg.Alertf("payment fail: cart(%d) place(%d) with err %+v", cart.ID, placeID, err)
			render.Respond(w, r, err)
			// TODO: do we return here?
			return
		}

		// check payment transaction
		if p.Transaction == nil {
			// something failed. Try again
			lg.Alertf("cart(%d) failed with empty transaction response", cart.ID)
			render.Respond(w, r, errors.New("payment failed, please try again"))
			return
		}

		if p.Transaction.Status != shopify.TransactionStatusSuccess {
			// something failed. Try again
			lg.Alertf("cart(%d) failed with transaction status %v", cart.ID, p.Transaction.Status)
			render.Respond(w, r, api.ErrCardVaultProcess(fmt.Errorf(p.Transaction.Message)))
			return
		}

		lg.Alertf("cart(%d) was just paid!", cart.ID)

		// 5. save shopify payment id
		sh.PaymentID = p.ID
		sh.PaymentDue = cc.PaymentDue
		sh.TotalTax = atoi(cc.TotalTax)
		sh.TotalPrice = atoi(cc.TotalPrice)
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
