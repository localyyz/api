package shopify

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
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
	ID     int64  `json:"id"`
	Amount string `json:"amount"`
	// clientside idempotency token
	UniqueToken                   string         `json:"unique_token"`
	PaymentProcessingErrorMessage string         `json:"payment_processing_error_message,omitempty"`
	PaymentToken                  *PaymentToken  `json:"payment_token"`
	RequestDetails                *RequestDetail `json:"request_details"`
}

type RequestDetail struct {
	IPAddress      string `json:"ip_address,omitempty"`
	AcceptLanguage string `json:"accept_language,omitempty"`
	UserAgent      string `json:"user_agent,omitempty"`
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
	var (
		resp                *http.Response
		pollWait            = "0"
		pollURL             = fmt.Sprintf("/admin/checkouts/%s/shipping_rates.json", token)
		pollStatus          = http.StatusAccepted
		shippingRateWrapper = new(ShippingRateRequest)
	)

	for {
		if pollStatus != http.StatusAccepted {
			break
		}

		req, err := c.client.NewRequest("GET", pollURL, nil)
		if err != nil {
			return nil, nil, err
		}

		wait, _ := strconv.Atoi(pollWait)
		// TODO: make a proper poller
		time.Sleep(time.Duration(wait) * time.Second)

		resp, err = c.client.Do(ctx, req, shippingRateWrapper)
		if err != nil {
			return nil, resp, err
		}

		// check Location and Retry-After for url and delay
		pollURL = resp.Header.Get("Location")
		pollWait = resp.Header.Get("Retry-After")
		pollStatus = resp.StatusCode
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
