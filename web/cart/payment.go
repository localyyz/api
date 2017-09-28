package cart

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/pressly/chi/render"
	"github.com/pressly/lg"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/lib/stripe"
	"bitbucket.org/moodie-app/moodie-api/web/api"
)

func ListPaymentMethods(w http.ResponseWriter, r *http.Request) {

}

type cartPaymentRequest struct {
	Number string `json:"number"`
	Type   string `json:"type"`
	Expiry string `json:"expiry"`
	Name   string `json:"name"`
	CVC    string `json:"cvc"`
}

func (c *cartPaymentRequest) Bind(r *http.Request) error {
	return nil
}

// TODO: Need to do the stripe token exchange on the frontend
// to be PCI complient. This is critical
func CreatePayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	//user := ctx.Value("session.user").(*data.User)
	cart := ctx.Value("cart").(*data.Cart)

	var payload cartPaymentRequest
	if err := render.Bind(r, &payload); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	expiryYM := strings.Split(payload.Expiry, "/")
	cardParam := &stripe.CardParams{
		Name:   payload.Name,
		Number: payload.Number,
		Month:  expiryYM[0],
		Year:   expiryYM[1],
		CVC:    payload.CVC,
	}

	// 1. exchange user credit card information for stripe token
	for placeID, sh := range cart.Etc.ShopifyData {
		ctx = context.WithValue(ctx, connect.StripeAccountKey, sh.PaymentAccountID)
		stripeToken, err := connect.ST.ExchangeToken(ctx, cardParam)
		lg.Warnf("%+v %+v", stripeToken, err)
		if err != nil {
			render.Render(w, r, api.ErrStripeProcess(err))
			return
		}

		// 2. send stripe payment token to shopify
		creds, err := data.DB.ShopifyCred.FindByPlaceID(placeID)
		if err != nil {
			render.Respond(w, r, err)
			return
		}

		api := shopify.NewClient(nil, creds.AccessToken)
		api.BaseURL, _ = url.Parse(creds.ApiURL)

		u, _ := uuid.NewUUID()
		payment := &shopify.PaymentRequest{
			Payment: &shopify.Payment{
				Amount:      sh.PaymentDue,
				UniqueToken: u.String(),
				PaymentToken: &shopify.PaymentToken{
					PaymentData: stripeToken.ID,
					Type:        shopify.StripeVaultToken,
				},
				RequestDetails: &shopify.RequestDetail{
					IPAddress:      stripeToken.ClientIP,
					AcceptLanguage: "EN",
				},
			},
		}
		_, _, err = api.Checkout.Payment(ctx, sh.Token, payment)
		if err != nil {
			render.Respond(w, r, err)
			return
		}
	}
	// TODO: create a customer on stripe after the first
	// tokenization so we can send stripe customer id moving forward
}
