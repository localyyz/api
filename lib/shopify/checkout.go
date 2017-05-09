package shopify

import (
	"context"
	"net/http"
)

type CheckoutService service

type Checkout struct {
	LineItems []*LineItem `json:"line_items,omitempty"`
	Email     string      `json:"email,omitempty"`
}

type BillingAddress struct{}

type LineItem struct {
	VariantId int64 `json:"variant_id"`
	Quantity  int64 `json:"quantity"`
}

type CheckoutRequest struct {
	Checkout *Checkout `json:"checkout"`
}

func (c *CheckoutService) Create(ctx context.Context, checkout *CheckoutRequest) (*Checkout, *http.Response, error) {
	req, err := c.client.NewRequest("POST", "/admin/checkouts.json", checkout)
	if err != nil {
		return nil, nil, err
	}

	cc := new(Checkout)
	resp, err := c.client.Do(ctx, req, cc)
	if err != nil {
		return nil, resp, err
	}

	return cc, resp, nil
}
