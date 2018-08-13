package apiclient

import (
	"context"
	"fmt"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/cart/express"
)

type ExpressCartService service

func (c *ExpressCartService) Get(ctx context.Context) (*presenter.Cart, *http.Response, error) {
	req, err := c.client.NewRequest("GET", "/carts/express", nil)
	if err != nil {
		return nil, nil, err
	}

	cartResponse := new(presenter.Cart)
	resp, err := c.client.Do(ctx, req, cartResponse)
	if err != nil {
		return nil, resp, err
	}

	return cartResponse, resp, nil
}

func (c *ExpressCartService) UpdateShippingAddress(ctx context.Context, address *data.CartAddress) (*presenter.Cart, *http.Response, error) {
	payload := express.ShippingAddressRequest{
		CartAddress: address,
		IsPartial:   false,
	}
	req, err := c.client.NewRequest("PUT", "/carts/express/shipping/address", payload)
	if err != nil {
		return nil, nil, err
	}

	cartResponse := new(presenter.Cart)
	resp, err := c.client.Do(ctx, req, cartResponse)
	if err != nil {
		return nil, resp, err
	}

	return cartResponse, resp, nil
}

func (c *ExpressCartService) UpdateShippingMethod(ctx context.Context, method string) (*presenter.Cart, *http.Response, error) {
	payload := express.ShippingMethodRequest{
		Handle: method,
	}
	req, err := c.client.NewRequest("PUT", "/carts/express/shipping/method", payload)
	if err != nil {
		return nil, nil, err
	}

	cartResponse := new(presenter.Cart)
	resp, err := c.client.Do(ctx, req, cartResponse)
	if err != nil {
		return nil, resp, err
	}

	return cartResponse, resp, nil
}

func (c *ExpressCartService) GetShippingRates(ctx context.Context) (*presenter.Cart, *http.Response, error) {
	req, err := c.client.NewRequest("GET", "/carts/express/shipping/estimate", nil)
	if err != nil {
		return nil, nil, err
	}

	cartResponse := new(presenter.Cart)
	resp, err := c.client.Do(ctx, req, cartResponse)
	if err != nil {
		return nil, resp, err
	}

	return cartResponse, resp, nil
}

func (c *ExpressCartService) Checkout(ctx context.Context) (*presenter.Cart, *http.Response, error) {
	req, err := c.client.NewRequest("POST", "/carts/express/checkout", nil)
	if err != nil {
		return nil, nil, err
	}

	cartResponse := new(presenter.Cart)
	resp, err := c.client.Do(ctx, req, cartResponse)
	if err != nil {
		return nil, resp, err
	}

	return cartResponse, resp, nil
}

func (c *ExpressCartService) AddDiscountCode(ctx context.Context, checkoutID int64, discountCode string) (*presenter.Checkout, *http.Response, error) {
	payload := struct {
		DiscountCode string `json:"discount"`
	}{discountCode}
	req, err := c.client.NewRequest("PUT", fmt.Sprintf("/carts/express/checkout/%d", checkoutID), payload)
	if err != nil {
		return nil, nil, err
	}
	checkoutResponse := new(presenter.Checkout)
	resp, err := c.client.Do(ctx, req, checkoutResponse)
	if err != nil {
		return nil, resp, err
	}
	return checkoutResponse, resp, nil
}

func (c *ExpressCartService) Pay(ctx context.Context, billing *data.CartAddress, token, email string) (*presenter.Cart, *http.Response, error) {
	payload := express.CartPaymentRequest{
		billing,
		token,
		email,
	}
	req, err := c.client.NewRequest("POST", "/carts/express/pay", payload)
	if err != nil {
		return nil, nil, err
	}

	cartResponse := new(presenter.Cart)
	resp, err := c.client.Do(ctx, req, cartResponse)
	if err != nil {
		return nil, resp, err
	}

	return cartResponse, resp, nil
}

func (c *ExpressCartService) AddItem(ctx context.Context, variant *data.ProductVariant) (*presenter.CartItem, *http.Response, error) {
	itemRequest := express.CartItemRequest{
		VariantID: &(variant.ID),
		ProductID: variant.ProductID,
		Color:     variant.Etc.Color,
		Size:      variant.Etc.Size,
	}
	req, err := c.client.NewRequest("POST", "/carts/express/items", itemRequest)
	if err != nil {
		return nil, nil, err
	}

	cartItemResponse := new(presenter.CartItem)
	resp, err := c.client.Do(ctx, req, cartItemResponse)
	if err != nil {
		return nil, resp, err
	}

	return cartItemResponse, resp, nil
}
