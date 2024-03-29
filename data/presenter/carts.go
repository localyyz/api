package presenter

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/render"
	"github.com/pressly/lg"
	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	xchange "bitbucket.org/moodie-app/moodie-api/lib/xchanger"
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

	TransactionID string `json:"transactionId,omitempty"`

	ctx context.Context
}

func (c *Cart) Render(w http.ResponseWriter, r *http.Request) error {
	if c.IsExpress {
		// if cart is express. pull values to the top
		for k, d := range c.Etc.ShopifyData {
			c.StripeAccountID = d.ShopifyPaymentAccountID
			c.Currency = d.Currency
			if s, ok := c.Etc.ShippingMethods[k]; ok && s != nil {
				c.TotalShipping += round(xchange.ToUSD(float64(s.Price/100.00), d.Currency))
			}
			c.TotalTax = round(xchange.ToUSD(float64(d.TotalTax/100.00), d.Currency))
			c.TotalPrice = round(xchange.ToUSD(float64(d.TotalPrice/100.00), d.Currency))
			if d.Discount != nil {
				c.TotalDiscount += atoi(d.Discount.Amount)
			}
			break
		}
		if rates, _ := c.ctx.Value("rates").([]*data.CartShippingMethod); rates != nil {
			for _, r := range rates {
				r.Price = round(xchange.ToUSD(float64(r.Price/100.00), c.Currency))
				c.ShippingRates = append(c.ShippingRates, r)
			}
		}
		if s := c.Etc.ShippingAddress; s != nil && c.ShippingAddress == nil {
			c.ShippingAddress = NewCartAddress(c.ctx, s)
			c.ShippingAddress.IsShipping = true
		}
		if b := c.Etc.BillingAddress; b != nil && c.BillingAddress == nil {
			c.BillingAddress = NewCartAddress(c.ctx, b)
			c.BillingAddress.IsBilling = true
		}
	} else {
		// NEW CHECKOUT
		for _, ch := range c.Checkouts {
			c.TotalShipping += round(xchange.ToUSD(ch.TotalShipping, ch.Currency))
			c.TotalTax += round(xchange.ToUSD(ch.TotalTax, ch.Currency))
			c.TotalPrice += round(xchange.ToUSD(ch.TotalPrice, ch.Currency))
			if ch.AppliedDiscount.AppliedDiscount != nil {
				c.TotalDiscount += atoi(ch.AppliedDiscount.Amount)
			}
		}

		// if not presented. present
		if s := c.Cart.ShippingAddress; s != nil && c.ShippingAddress == nil {
			c.ShippingAddress = NewCartAddress(c.ctx, s)
			c.ShippingAddress.IsShipping = true
		}
		// if not presented. present
		if b := c.Cart.BillingAddress; b != nil && c.BillingAddress == nil {
			c.BillingAddress = NewCartAddress(c.ctx, b)
			c.BillingAddress.IsBilling = true
		}
	}

	for _, ci := range c.CartItems {
		ci.Render(w, r)
	}

	// NOTE MAKE THIS BETTER ... check deal price on cart items
	if !c.IsExpress && c.Status == data.CartStatusInProgress && len(c.Checkouts) > 0 {
		// find appliable discounts... productID -> value
		discounts := map[int64]float64{}
		for _, ch := range c.Checkouts {
			if ch.AppliedDiscount != nil && ch.AppliedDiscount.AppliedDiscount != nil {
				// NOTE: this is dirty as fuck. make this better
				deals, _ := data.DB.Deal.FindAll(db.Cond{
					"merchant_id": ch.PlaceID,
					"code":        ch.AppliedDiscount.AppliedDiscount.Title,
					"status":      data.DealStatusActive,
				})
				place, err := data.DB.Place.FindByID(ch.PlaceID)
				if err != nil {
					lg.Alert("failed to fetch place(%d) for checkout discount code: %v", place.ID, err)
					continue
				}
				for _, d := range deals {
					prs, _ := data.DB.DealProduct.FindByDealID(d.ID)
					for _, p := range prs {
						discounts[p.ProductID] = xchange.ToUSD(d.Value, place.Currency)
					}
				}
			}
		}
		for _, item := range c.CartItems {
			if v, ok := discounts[item.ProductID]; ok {
				item.Price += v
			}
		}
	}

	return nil
}

func NewCart(ctx context.Context, cart *data.Cart) *Cart {
	resp := &Cart{
		Cart: cart,
		ctx:  ctx,
	}

	dbItems, err := data.DB.CartItem.FindByCartID(cart.ID)
	if err != nil {
		return resp
	}
	resp.ShippingAddress = NewCartAddress(ctx, cart.ShippingAddress)
	resp.ShippingAddress.IsShipping = true

	resp.BillingAddress = NewCartAddress(ctx, cart.BillingAddress)
	resp.BillingAddress.IsBilling = true

	if checkouts, _ := data.DB.Checkout.FindAllByCartID(cart.ID); len(checkouts) > 0 {
		for _, c := range checkouts {
			resp.Checkouts = append(resp.Checkouts, NewCheckout(ctx, c))
		}
	}

	resp.CartItems = make(CartItemList, 0)
	for _, item := range dbItems {
		ci := NewCartItem(ctx, item)
		resp.CartItems = append(resp.CartItems, ci)
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
