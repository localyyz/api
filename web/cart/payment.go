package cart

import (
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/goware/lg"
	"github.com/pressly/chi/render"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/lib/stripe"
	"bitbucket.org/moodie-app/moodie-api/web/api"
)

type cartPaymentRequest struct {
	Number string `json:"number"`
	Year   string `json:"exp_year"`
	Month  string `json:"exp_month"`
	CVC    string `json:"cvc"`
}

func (c *cartPaymentRequest) Bind(r *http.Request) error {
	return nil
}

func Payment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	//user := ctx.Value("session.user").(*data.User)
	cart := ctx.Value("cart").(*data.Cart)

	var payload cartPaymentRequest
	if err := render.Bind(r, &payload); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	cardParam := &stripe.CardParams{
		Number: payload.Number,
		Year:   payload.Year,
		Month:  payload.Month,
		CVC:    payload.CVC,
	}

	// 1. exchange user credit card information for stripe token
	for placeID, sh := range cart.Etc.ShopifyData {
		stripeToken, err := connect.ST.ExchangeToken(ctx, sh.PaymentAccountID, cardParam)
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
					UserAgent:      api.UserAgent,
				},
			},
		}
		p, _, err := api.Checkout.Payment(ctx, sh.Token, payment)
		if err != nil {
			render.Respond(w, r, err)
			return
		}
		lg.Warn(p)
	}

	// TODO: create a customer on stripe after the first
	// tokenization so we can send stripe customer id moving forward
}
