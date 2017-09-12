package shopify

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type CheckoutService service

type Checkout struct {
	*CheckoutPrice // embed

	LineItems        []*LineItem `json:"line_items,omitempty"`
	Email            string      `json:"email,omitempty"`
	Token            string      `json:"token,omitempty"`
	Name             string      `json:"name,omitempty"`
	CustomerID       int64       `json:"customer_id,omitempty"`
	PaymentAccountID string      `json:"shopify_payments_account_id,omitempty"`
	WebURL           string      `json:"web_url,omitempty"`
	WebProcessingURL string      `json:"web_processing_url,omitempty"`

	ShippingAddress *CustomerAddress `json:"shipping_address,omitempty"`
	BillingAddress  *CustomerAddress `json:"billing_address,omitempty"`
	ShippingLine    *ShippingLine    `json:"shipping_line,omitempty"`
}

type ShippingLine struct {
	Handle string `json:"handle,omitempty"`
	Price  string `json:"price,omitempty"`
	Title  string `json:"title,omitempty"`
}

type CheckoutShipping struct {
	ID            string        `json:"id"`
	Price         string        `json:"price"`
	Title         string        `json:"title"`
	Checkout      *ShippingRate `json:"checkout"`
	PhoneRequired bool          `json:"phone_required"`
	DeliveryRange []time.Time   `json:"delivery_range"`
	Handle        string        `json:"handle"`
}

type ShippingRate struct {
	SubtotalPrice string `json:"subtotal_price"`
	TotalTax      string `json:"total_tax"`
	TotalPrice    string `json:"total_price"`
	PaymentDue    string `json:"payment_due"`
}
type CheckoutPrice ShippingRate

type BillingAddress struct{}

type LineItem struct {
	VariantID int64 `json:"variant_id"`
	Quantity  int64 `json:"quantity"`
}

type CheckoutRequest struct {
	Checkout *Checkout `json:"checkout"`
}

type ShippingRateRequest struct {
	CheckoutShipping []*CheckoutShipping `json:"shipping_rates"`
}

type Payment struct {
	Amount string `json:"amount"`
	// clientside idempotency token
	UniqueToken    string         `json:"unique_token"`
	PaymentToken   *PaymentToken  `json:"payment_token"`
	RequestDetails *RequestDetail `json:"request_details"`
}

type RequestDetail struct {
	IPAddress      string `json:"ip_address"`
	AcceptLanguage string `json:"accept_language"`
	UserAgent      string `json:"user_agent"`
}

type PaymentToken struct {
	// Stripe token
	PaymentData string `json:"payment_data"`
	// stripe_vault_token
	Type string `json:"type"`
}

type PaymentRequest struct {
	Payment *Payment `json:"payment"`
}

const StripeVaultToken = `stripe_vault_token`

func (c *CheckoutService) Create(ctx context.Context, checkout *CheckoutRequest) (*Checkout, *http.Response, error) {
	req, err := c.client.NewRequest("POST", "/admin/checkouts.json", checkout)
	if err != nil {
		return nil, nil, err
	}

	checkoutWrapper := new(CheckoutRequest)
	resp, err := c.client.Do(ctx, req, checkoutWrapper)
	if err != nil {
		return nil, resp, err
	}

	return checkoutWrapper.Checkout, resp, nil
}

func (c *CheckoutService) Update(ctx context.Context, checkout *CheckoutRequest) (*Checkout, *http.Response, error) {
	req, err := c.client.NewRequest("PUT", fmt.Sprintf("/admin/checkouts/%s.json", checkout.Checkout.Token), checkout)
	if err != nil {
		return nil, nil, err
	}

	checkoutWrapper := new(CheckoutRequest)
	resp, err := c.client.Do(ctx, req, checkoutWrapper)
	if err != nil {
		return nil, resp, err
	}

	return checkoutWrapper.Checkout, resp, nil
}

func (c *CheckoutService) ListShippingRates(ctx context.Context, token string) ([]*CheckoutShipping, *http.Response, error) {
	req, err := c.client.NewRequest("GET", fmt.Sprintf("/admin/checkouts/%s/shipping_rates.json", token), nil)
	if err != nil {
		return nil, nil, err
	}

	shippingRateWrapper := new(ShippingRateRequest)
	resp, err := c.client.Do(ctx, req, shippingRateWrapper)
	if err != nil {
		return nil, resp, err
	}

	return shippingRateWrapper.CheckoutShipping, resp, nil
}

func (c *CheckoutService) Payment(ctx context.Context, token string, payment *PaymentRequest) (*Payment, *http.Response, error) {
	req, err := c.client.NewRequest("POST", fmt.Sprintf("/admin/checkouts/%s/payments.json", token), payment)
	if err != nil {
		return nil, nil, err
	}

	paymentWrapper := new(PaymentRequest)
	resp, err := c.client.Do(ctx, req, paymentWrapper)
	if err != nil {
		return nil, resp, err
	}

	return paymentWrapper.Payment, resp, nil

}
