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

	CartItems CartItemList `json:"items"`

	TotalShipping int64 `json:"totalShipping"`
	TotalTax      int64 `json:"totalTax"`
	TotalPrice    int64 `json:"totalPrice"`
	TotalDiscount int64 `json:"totalDiscount"`

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

	// calculate cart subtotal by line item
	if cart.Etc.ShopifyData != nil {
		for k, d := range cart.Etc.ShopifyData {
			resp.TotalTax += d.TotalTax
			resp.TotalPrice += d.TotalPrice
			if d.Discount != nil {
				resp.TotalDiscount += atoi(d.Discount.Amount)
			}
			if s, ok := cart.Etc.ShippingMethods[k]; ok && s != nil {
				resp.TotalShipping += s.Price
			}
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
