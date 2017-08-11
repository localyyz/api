package presenter

import (
	"context"
	"net/http"

	"github.com/pressly/chi/render"

	"bitbucket.org/moodie-app/moodie-api/data"
)

type Cart struct {
	*data.Cart

	CartItems     CartItemList               `json:"items"`
	ShippingRates []*data.CartShippingMethod `json:"shippingRates,omitempty"`

	Subtotal float64 `json:"subtotal"` // in cents

	ctx context.Context
}

func (c *Cart) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func NewCart(ctx context.Context, cart *data.Cart) *Cart {
	resp := &Cart{
		Cart: cart,
		ctx:  ctx,
	}

	resp.CartItems = make(CartItemList, 0, 0)
	if dbItems, _ := data.DB.CartItem.FindByCartID(cart.ID); dbItems != nil {
		var cartItems CartItemList
		for _, item := range dbItems {
			cartItems = append(cartItems, NewCartItem(ctx, item))
		}
		resp.CartItems = cartItems
	}

	resp.ShippingRates = make([]*data.CartShippingMethod, 0)
	if rates, _ := ctx.Value("shipping.rates").([]*data.CartShippingMethod); rates != nil {
		resp.ShippingRates = rates
	}

	// calculate cart subtotal by line item
	for _, cartItem := range resp.CartItems {
		resp.Subtotal += (cartItem.Price * 100.0)
	}

	return resp
}

type UserCartList []*Cart

func (l UserCartList) Render(w http.ResponseWriter, r *http.Request) error {
	for _, v := range l {
		if err := v.Render(w, r); err != nil {
			return err
		}
	}
	return nil
}

func NewUserCartList(ctx context.Context, carts []*data.Cart) []render.Renderer {
	list := []render.Renderer{}
	for _, cart := range carts {
		list = append(list, NewCart(ctx, cart))
	}
	return list
}
