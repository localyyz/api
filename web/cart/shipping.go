package cart

import (
	"context"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/goware/lg"
	"github.com/pkg/errors"
	"github.com/pressly/chi/render"
	db "upper.io/db.v3"
)

func ListShippingMethods(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cart := ctx.Value("cart").(*data.Cart)

	var placeIDs []int64
	tokensMap := map[int64]string{}
	for ID, d := range cart.Etc.ShopifyData {
		placeIDs = append(placeIDs, ID)
		tokensMap[ID] = d.Token
	}
	creds, err := data.DB.ShopifyCred.FindAll(db.Cond{"place_id": placeIDs})
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	ratesMap := map[string]*data.CartShippingMethod{}
	for _, cred := range creds {
		api := shopify.NewClient(nil, cred.AccessToken)
		api.BaseURL, _ = url.Parse(cred.ApiURL)

		shippingMethods, _, _ := api.Checkout.ListShippingRates(ctx, tokensMap[cred.PlaceID])
		// TODO/NOTE: there is a weird shopify bug that first call to this api
		// endpoint will always result in empty response. try again until
		// something is back
		maxAttempt := 2
		attempt := 1
		for {
			if attempt == maxAttempt {
				break
			}
			if len(shippingMethods) == 0 {
				shippingMethods, _, err = api.Checkout.ListShippingRates(ctx, tokensMap[cred.PlaceID])
				if err != nil {
					break
				}
				time.Sleep(time.Second)

				attempt += 1
				continue
			}
			break
		}

		for _, m := range shippingMethods {
			r, ok := ratesMap[m.ID]
			if ok {
				r.Price += atof(m.Price)
				continue
			}
			ratesMap[m.ID] = &data.CartShippingMethod{
				Handle:        m.ID,
				Title:         m.Title,
				Price:         atof(m.Price),
				DeliveryRange: m.DeliveryRange,
			}
		}
	}

	rates := []*data.CartShippingMethod{}
	for _, m := range ratesMap {
		rates = append(rates, m)
	}
	sort.Slice(rates, func(i, j int) bool { return rates[i].Price < rates[j].Price })
	ctx = context.WithValue(ctx, "shipping.rates", rates)

	render.Respond(w, r, presenter.NewCart(ctx, cart))
}

func atof(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		lg.Errorf("failed to parse %s to float", s)
		return f
	}
	return f * 100.0
}

type cartUpdateShippingRequest struct {
	ShippingMethod *data.CartShippingMethod `json:"shippingMethod"`
}

func (c *cartUpdateShippingRequest) Bind(r *http.Request) error {
	return nil
}

func UpdateShippingMethod(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cart := ctx.Value("cart").(*data.Cart)

	var payload cartUpdateShippingRequest
	if err := render.Bind(r, &payload); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	// sync to shopify
	var placeIDs []int64
	tokensMap := map[int64]string{}
	for ID, d := range cart.Etc.ShopifyData {
		placeIDs = append(placeIDs, ID)
		tokensMap[ID] = d.Token
	}
	creds, err := data.DB.ShopifyCred.FindAll(db.Cond{"place_id": placeIDs})
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	cart.Etc.ShippingMethods = make(map[int64]*data.CartShippingMethod)
	for _, cred := range creds {
		api := shopify.NewClient(nil, cred.AccessToken)
		api.BaseURL, _ = url.Parse(cred.ApiURL)

		checkout := &shopify.Checkout{
			Token: tokensMap[cred.PlaceID],
			ShippingLine: &shopify.ShippingLine{
				Handle: payload.ShippingMethod.Handle,
			},
		}
		c, _, err := api.Checkout.Update(ctx, &shopify.CheckoutRequest{checkout})
		if err != nil {
			lg.Alert(errors.Wrapf(err, "checkout shipping update. cart(%d)", cart.ID))
			continue
		}
		// TODO handle error here and should retry

		cart.Etc.ShippingMethods[cred.PlaceID] = &data.CartShippingMethod{
			Handle: payload.ShippingMethod.Handle,
			Price:  atof(c.ShippingLine.Price),
			Title:  c.ShippingLine.Title,
		}
		cart.Etc.ShopifyData[cred.PlaceID].SubtotalPrice = atof(c.SubtotalPrice)
		cart.Etc.ShopifyData[cred.PlaceID].TotalPrice = atof(c.TotalPrice)
		cart.Etc.ShopifyData[cred.PlaceID].TotalTax = atof(c.TotalTax)
		cart.Etc.ShopifyData[cred.PlaceID].PaymentDue = c.PaymentDue
	}

	if err := data.DB.Cart.Save(cart); err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Render(w, r, presenter.NewCart(ctx, cart))
}