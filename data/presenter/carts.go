package presenter

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/render"
	"github.com/pressly/lg"

	"bitbucket.org/moodie-app/moodie-api/data"
)

type Cart struct {
	*data.Cart

	// For express carts quick access values
	Currency        string `json:"currency,omitempty"`
	StripeAccountID string `json:"stripeAccountId,omitempty"`

	CartItems       CartItemList `json:"items"`
	Checkouts       []*Checkout  `json:"checkouts"`
	ShippingAddress *CartAddress `json:"shippingAddress"`
	BillingAddress  *CartAddress `json:"billingAddress"`

	HasError  bool   `json:"hasError"`
	Error     string `json:"error"`
	ErrorCode uint32 `json:"errorCode"`

	ctx context.Context
}

func (c *Cart) Render(w http.ResponseWriter, r *http.Request) error {
	if c.IsExpress {
		// if cart is express. pull values to the top
		for _, d := range c.Etc.ShopifyData {
			c.StripeAccountID = d.ShopifyPaymentAccountID
			c.Currency = d.Currency
			break
		}
	}

	return nil
}

func NewCart(ctx context.Context, cart *data.Cart) *Cart {
	resp := &Cart{
		Cart: cart,
		ctx:  ctx,
	}

	resp.CartItems = make(CartItemList, 0, 0)
	dbItems, err := data.DB.CartItem.FindByCartID(cart.ID)
	if err != nil {
		lg.Warn(err)
		return resp
	}

	var cartItems CartItemList
	for _, item := range dbItems {
		cartItems = append(cartItems, NewCartItem(ctx, item))
	}
	resp.CartItems = cartItems

	resp.ShippingAddress = &CartAddress{IsShipping: true}
	resp.BillingAddress = &CartAddress{IsBilling: true}
	if s := cart.ShippingAddress; s != nil {
		resp.ShippingAddress.CartAddress = s
	} else if s := cart.Etc.ShippingAddress; s != nil {
		resp.ShippingAddress.CartAddress = s
	}

	if b := cart.BillingAddress; b != nil {
		resp.BillingAddress.CartAddress = b
	} else if b := cart.Etc.BillingAddress; b != nil {
		resp.BillingAddress.CartAddress = b
	}

	if checkouts, _ := data.DB.Checkout.FindAllByCartID(cart.ID); len(checkouts) > 0 {
		for _, c := range checkouts {
			resp.Checkouts = append(resp.Checkouts, NewCheckout(ctx, c))
		}
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

func atoi(s string) int64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		lg.Errorf("failed to parse %s to float", s)
		return 0
	}
	return int64(f * 100.0)
}
