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

	CartItems       CartItemList `json:"items"`
	Checkouts       []*Checkout  `json:"checkouts"`
	ShippingAddress *CartAddress `json:"shippingAddress"`
	BillingAddress  *CartAddress `json:"billingAddress"`

	HasError  bool   `json:"hasError"`
	Error     string `json:"error"`
	ErrorCode uint32 `json:"errorCode"`

	// NOTE: backwards compatible
	ShippingRates   []*data.CartShippingMethod `json:"shippingRates"`
	Currency        string                     `json:"currency,omitempty"`
	StripeAccountID string                     `json:"stripeAccountId,omitempty"`
	TotalShipping   int64                      `json:"totalShipping"`
	TotalTax        int64                      `json:"totalTax"`
	TotalPrice      int64                      `json:"totalPrice"`
	TotalDiscount   int64                      `json:"totalDiscount"`

	ctx context.Context
}

func (c *Cart) Render(w http.ResponseWriter, r *http.Request) error {
	if c.IsExpress {
		// if cart is express. pull values to the top
		for k, d := range c.Etc.ShopifyData {
			c.StripeAccountID = d.ShopifyPaymentAccountID
			c.Currency = d.Currency
			if s, ok := c.Etc.ShippingMethods[k]; ok && s != nil {
				c.TotalShipping += s.Price
			}
			c.TotalTax = d.TotalTax
			c.TotalPrice = d.TotalPrice
			if d.Discount != nil {
				c.TotalDiscount += atoi(d.Discount.Amount)
			}
			break
		}
		if rates, _ := c.ctx.Value("rates").([]*data.CartShippingMethod); rates != nil {
			c.ShippingRates = rates
		}
		if s := c.Etc.ShippingAddress; s != nil {
			c.ShippingAddress = NewCartAddress(c.ctx, s)
			c.ShippingAddress.IsShipping = true
		}
		if b := c.Etc.BillingAddress; b != nil {
			c.BillingAddress = NewCartAddress(c.ctx, b)
			c.BillingAddress.IsBilling = true
		}
	} else {
		// NEW CHECKOUT
		for _, ch := range c.Checkouts {
			c.TotalShipping += round(ch.TotalShipping)
			c.TotalTax += round(ch.TotalTax)
			c.TotalPrice += round(ch.TotalPrice)

			if ch.AppliedDiscount.AppliedDiscount != nil {
				c.TotalDiscount += atoi(ch.AppliedDiscount.Amount)
			}
		}

		if s := c.Cart.ShippingAddress; s != nil {
			c.ShippingAddress = NewCartAddress(c.ctx, s)
			c.ShippingAddress.IsShipping = true
		}
		if b := c.Cart.BillingAddress; b != nil {
			c.BillingAddress = NewCartAddress(c.ctx, b)
			c.BillingAddress.IsBilling = true
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

	resp.ShippingAddress = NewCartAddress(ctx, cart.ShippingAddress)
	resp.ShippingAddress.IsShipping = true

	resp.BillingAddress = NewCartAddress(ctx, cart.BillingAddress)
	resp.BillingAddress.IsBilling = true

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

func round(f float64) int64 {
	return int64(f * 100.0)
}

func atoi(s string) int64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		lg.Errorf("failed to parse %s to float", s)
		return 0
	}
	return round(f)
}
