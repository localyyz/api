package shopify

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type CheckoutService service

type Checkout struct {
	LineItems        []*LineItem `json:"line_items,omitempty"`
	Email            string      `json:"email,omitempty"`
	Token            string      `json:"token,omitempty"`
	Name             string      `json:"name,omitempty"`
	CustomerID       int64       `json:"customer_id,omitempty"`
	PaymentAccountID string      `json:"shopify_payments_account_id,omitempty"`
	WebURL           string      `json:"web_url,omitempty"`
	WebProcessingURL string      `json:"web_processing_url,omitempty"`

	*ShippingRate // embed

	ShippingAddress *CustomerAddress `json:"shipping_address,omitempty"`
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
	TotalTax      string `json:"total_tax"`
	TotalPrice    string `json:"total_price"`
	SubtotalPrice string `json:"subtotal_price"`
}

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
