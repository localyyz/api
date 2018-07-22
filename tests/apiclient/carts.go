package apiclient

import (
	"context"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/shopper"
	"bitbucket.org/moodie-app/moodie-api/web/cart/cartitem"
)

type CartService service

func (c *CartService) Get(ctx context.Context) (*presenter.Cart, *http.Response, error) {
	req, err := c.client.NewRequest("GET", "/carts/default", nil)
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

func (c *CartService) Put(ctx context.Context, cart *data.Cart) (*presenter.Cart, *http.Response, error) {
	putRequest := struct {
		ShippingAddress *data.CartAddress `json:"shippingAddress"`
		BillingAddress  *data.CartAddress `json:"billingAddress"`
		Email           string            `json:"email"`
	}{
		cart.ShippingAddress,
		cart.BillingAddress,
		cart.Email,
	}

	req, err := c.client.NewRequest("PUT", "/carts/default", putRequest)
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

func (c *CartService) Checkout(ctx context.Context) (*presenter.Cart, *http.Response, error) {
	req, err := c.client.NewRequest("POST", "/carts/default/checkout", nil)
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

func (c *CartService) Pay(ctx context.Context, card *shopper.PaymentCard) (*presenter.Cart, *http.Response, error) {
	payRequest := struct {
		Card *shopper.PaymentCard `json:"payment"`
	}{card}
	req, err := c.client.NewRequest("POST", "/carts/default/pay", payRequest)
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

func (c *CartService) AddItem(ctx context.Context, item *data.CartItem) (*presenter.CartItem, *http.Response, error) {
	itemRequest := cartitem.CartItemRequest{
		VariantID: &(item.VariantID),
	}
	req, err := c.client.NewRequest("POST", "/carts/default/items", itemRequest)
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
