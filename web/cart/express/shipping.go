package express

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"github.com/go-chi/render"
	"github.com/pkg/errors"
)

func GetShippingRates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cart := ctx.Value("cart").(*data.Cart)
	client := ctx.Value("shopify.client").(*shopify.Client)
	checkout := ctx.Value("shopify.checkout").(*data.CartShopifyData)

	shopifyRates, _, _ := client.Checkout.ListShippingRates(ctx, checkout.Token)
	var rates []*data.CartShippingMethod
	for _, r := range shopifyRates {
		rates = append(
			rates,
			&data.CartShippingMethod{
				Handle:        r.Handle,
				Price:         atoi(r.Price),
				Title:         r.Title,
				DeliveryRange: r.DeliveryRange,
			},
		)
	}
	ctx = context.WithValue(ctx, "rates", rates)
	render.Render(w, r, presenter.NewCart(ctx, cart))
}

type shippingAddressRequest struct {
	*data.CartAddress
}

func (p *shippingAddressRequest) Bind(r *http.Request) error {
	if p.CountryCode == "" {
		return errors.New("shipping address missing country")
	}

	// NOTE: the address passed in here could be truncated (apple pay privacy)
	// append mock data
	if len(p.Address) == 0 {
		p.Address = "1 Mock Street"
	}

	switch strings.ToLower(p.CountryCode) {
	case "ca":
		// Canada postal code is truncated. Use placeholder for last three
		// characters
		p.Zip = fmt.Sprintf("%s 9Z0", p.Zip)
	case "uk":
		// TODO
	}

	if len(p.FirstName) == 0 {
		p.FirstName = "Johnny"
	}
	if len(p.LastName) == 0 {
		p.LastName = "Appleseed"
	}
	return nil
}

func UpdateShippingAddress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	client := ctx.Value("shopify.client").(*shopify.Client)
	checkout := ctx.Value("shopify.checkout").(*data.CartShopifyData)
	cart := ctx.Value("cart").(*data.Cart)

	var payload shippingAddressRequest
	if err := render.Bind(r, &payload); err != nil {
		render.Respond(w, r, err)
		return
	}

	ch, _, err := client.Checkout.Update(
		ctx,
		&shopify.CheckoutRequest{
			Checkout: &shopify.Checkout{
				// Partial customer address from apple pay. enough to get shipping rates
				Token: checkout.Token,
				ShippingAddress: &shopify.CustomerAddress{
					FirstName:    payload.FirstName,
					LastName:     payload.LastName,
					Country:      payload.Country,
					CountryCode:  payload.CountryCode,
					Province:     payload.Province,
					ProvinceCode: payload.ProvinceCode,
					City:         payload.City,
					Address1:     payload.Address,
					Zip:          payload.Zip,
				},
			},
		},
	)
	if err != nil {
		render.Respond(w, r, errors.Wrap(err, "express cart shipping rate"))
		return
	}

	// TODO return tax_lines for more detailed tax breakdown
	checkout.TotalTax = atoi(ch.TotalTax)
	checkout.TotalPrice = atoi(ch.TotalPrice)
	checkout.PaymentDue = ch.PaymentDue
	cart.Etc.ShopifyData[checkout.PlaceID] = checkout
	// shipping address has changed -> reset shipping method
	cart.Etc.ShippingMethods = make(map[int64]*data.CartShippingMethod)

	cart.Etc.ShippingAddress = payload.CartAddress
	data.DB.Cart.Save(cart)

	render.Render(w, r, presenter.NewCart(ctx, cart))
}

type shippingMethodRequest struct {
	Handle string `json:"handle"`
}

func (*shippingMethodRequest) Bind(r *http.Request) error {
	return nil
}

func UpdateShippingMethod(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	client := ctx.Value("shopify.client").(*shopify.Client)
	checkout := ctx.Value("shopify.checkout").(*data.CartShopifyData)
	cart := ctx.Value("cart").(*data.Cart)

	var payload shippingMethodRequest
	if err := render.Bind(r, &payload); err != nil {
		render.Respond(w, r, err)
		return
	}

	ch, _, err := client.Checkout.Update(
		ctx,
		&shopify.CheckoutRequest{
			Checkout: &shopify.Checkout{
				Token: checkout.Token,
				ShippingLine: &shopify.ShippingLine{
					Handle: payload.Handle,
				},
			},
		},
	)
	if err != nil {
		render.Respond(w, r, errors.Wrap(err, "express cart shipping rate"))
		return
	}

	if cart.Etc.ShippingMethods == nil {
		cart.Etc.ShippingMethods = make(map[int64]*data.CartShippingMethod)
	}
	// save shippng method to db
	cart.Etc.ShippingMethods[checkout.PlaceID] = &data.CartShippingMethod{
		Handle: ch.ShippingLine.Handle,
		Price:  atoi(ch.ShippingLine.Price),
		Title:  ch.ShippingLine.Title,
	}
	checkout.TotalTax = atoi(ch.TotalTax)
	checkout.TotalPrice = atoi(ch.TotalPrice)
	checkout.PaymentDue = ch.PaymentDue
	cart.Etc.ShopifyData[checkout.PlaceID] = checkout

	if err := data.DB.Save(cart); err != nil {
		render.Respond(w, r, err)
		return
	}
	render.Render(w, r, presenter.NewCart(ctx, cart))
}
