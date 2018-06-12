package shopper

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/google/uuid"
	db "upper.io/db.v3"
)

type PaymentCard struct {
	Number string `json:"number"`
	Type   string `json:"type"`
	Expiry string `json:"expiry"`
	Name   string `json:"name"`
	CVC    string `json:"cvc"`
}

func (c *PaymentCard) Bind(r *http.Request) error {
	return nil
}

type Payment struct {
	uniqueID       string
	checkout       *data.Checkout
	card           *PaymentCard
	billingAddress *data.CartAddress
	requestIP      string
	authToken      string

	Err error
}

type Payer interface {
	ExchangeToken(context.Context, *shopify.Payment) error
	HandleError(*shopify.Payment, error)
	Finalize(*shopify.Payment) error
	Do(ctx context.Context) error
}

var _ interface {
	Payer
} = &Payment{}

func NewPayment(ctx context.Context, checkout *data.Checkout) *Payment {
	billingAddress := ctx.Value(BillingAddressCtxKey).(*data.CartAddress)
	requestIP := ctx.Value(RequestIPCtxKey).(string)
	card := ctx.Value(PaymentCardCtxKey).(*PaymentCard)

	u, _ := uuid.NewUUID()
	return &Payment{
		uniqueID:       u.String(),
		checkout:       checkout,
		card:           card,
		billingAddress: billingAddress,
		requestIP:      requestIP,
	}
}

func (p *Payment) ExchangeToken(ctx context.Context, req *shopify.Payment) error {
	var (
		authToken string
		err       error
	)
	if len(p.checkout.PaymentAccountID) != 0 {
		authToken, err = NewStripeToken(ctx, p)
		req.PaymentToken = &shopify.PaymentToken{
			PaymentData: authToken,
			Type:        shopify.StripeVaultToken,
		}
	} else {
		authToken, err = NewVaultToken(ctx, p)
		req.SessionID = authToken
	}

	if err != nil {
		return err
	}

	// do something with auth token
	return nil
}

func (p *Payment) HandleError(req *shopify.Payment, err error) {
	if err != nil {
		p.Err = err
		return
	}
	if t := req.Transaction; t != nil {
		if t.Status != shopify.TransactionStatusSuccess {
			p.Err = api.ErrCardVaultProcess(
				fmt.Errorf("%s (Code: %s, Status: %v)", t.Message, t.ErrorCode, t.Status),
			)
		}
	}
}

func (p *Payment) Finalize(req *shopify.Payment) error {
	if p.Err != nil {
		p.checkout.Status = data.CheckoutStatusPaymentFailed
	} else {
		p.checkout.SuccessPaymentID = req.Transaction.ID
		p.checkout.Status = data.CheckoutStatusPaymentSuccess
	}
	if err := data.DB.Checkout.Save(p.checkout); err != nil {
		return err
	}
	return p.Err
}

func (p *Payment) Do(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	req := &shopify.Payment{
		Amount:      p.checkout.PaymentDue,
		UniqueToken: p.uniqueID,
	}
	// fetch payment token from Stripe or CardVault
	if err := p.ExchangeToken(ctx, req); err != nil {
		return err
	}

	// update request detail, requestIp may have changed
	req.RequestDetails = &shopify.RequestDetail{
		IPAddress: p.requestIP,
	}

	p.HandleError(p.doPayment(ctx, req))

	return p.Finalize(req)
}

func (p *Payment) doPayment(ctx context.Context, req *shopify.Payment) (*shopify.Payment, error) {
	cred, err := data.DB.ShopifyCred.FindOne(db.Cond{"place_id": p.checkout.PlaceID})
	if err != nil {
		return req, err
	}

	client := shopify.NewClient(nil, cred.AccessToken)
	client.BaseURL, _ = url.Parse(cred.ApiURL)
	client.Debug = true

	_, _, err = client.Checkout.Payment(ctx, p.checkout.Token, req)
	return req, err
}
