package cart

import (
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

type shopRate struct {
	PlaceID int64                      `json:"placeId"`
	Rates   []*data.CartShippingMethod `json:"rates"`
}

func ListShippingRates(w http.ResponseWriter, r *http.Request) {
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

	shopRates := make([]*shopRate, len(creds))
	for i, cred := range creds {
		api := shopify.NewClient(nil, cred.AccessToken)
		api.BaseURL, _ = url.Parse(cred.ApiURL)

		m, _, _ := api.Checkout.ListShippingRates(ctx, tokensMap[cred.PlaceID])
		// TODO/NOTE: there is a weird shopify bug that first call to this api
		// endpoint will always result in empty response. try again until
		// something is back
		maxAttempt := 2
		attempt := 1
		for {
			if attempt == maxAttempt {
				break
			}
			if len(m) == 0 {
				m, _, err = api.Checkout.ListShippingRates(ctx, tokensMap[cred.PlaceID])
				if err != nil {
					break
				}
				time.Sleep(time.Second)

				attempt += 1
				continue
			}
			break
		}

		rates := make([]*data.CartShippingMethod, len(m))
		for ii, mm := range m {
			rates[ii] = &data.CartShippingMethod{
				Handle:        mm.ID,
				Title:         mm.Title,
				Price:         atoi(mm.Price),
				DeliveryRange: mm.DeliveryRange,
			}
		}
		sort.Slice(rates, func(i, j int) bool { return rates[i].Price < rates[j].Price })
		shopRates[i] = &shopRate{
			PlaceID: cred.PlaceID,
			Rates:   rates,
		}
	}
	render.Respond(w, r, shopRates)
}

func atoi(s string) int64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		lg.Errorf("failed to parse %s to float", s)
		return 0
	}
	return int64(f * 100.0)
}

type cartUpdateShippingRequest struct {
	ShippingMethods map[int64]*data.CartShippingMethod `json:"shippingMethods"`
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

	// TODO: handle cases where different shipping methods are passed in.
	// for now, select the first one
	var method *data.CartShippingMethod
	for _, v := range payload.ShippingMethods {
		method = v
		break
	}

	//cart.Etc.ShippingMethods = make(map[int64]*data.CartShippingMethod)
	for _, cred := range creds {
		api := shopify.NewClient(nil, cred.AccessToken)
		api.BaseURL, _ = url.Parse(cred.ApiURL)

		checkout := &shopify.Checkout{
			Token: tokensMap[cred.PlaceID],
			ShippingLine: &shopify.ShippingLine{
				Handle: method.Handle,
			},
		}
		c, _, err := api.Checkout.Update(ctx, &shopify.CheckoutRequest{checkout})
		if err != nil {
			lg.Alert(errors.Wrapf(err, "checkout shipping update. cart(%d)", cart.ID))
			continue
		}
		// TODO handle error here and should retry

		cart.Etc.ShippingMethods[cred.PlaceID] = method
		cart.Etc.ShopifyData[cred.PlaceID].SubtotalPrice = atoi(c.SubtotalPrice)
		cart.Etc.ShopifyData[cred.PlaceID].TotalPrice = atoi(c.TotalPrice)
		cart.Etc.ShopifyData[cred.PlaceID].TotalTax = atoi(c.TotalTax)
		cart.Etc.ShopifyData[cred.PlaceID].PaymentDue = c.PaymentDue
	}

	if err := data.DB.Cart.Save(cart); err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Render(w, r, presenter.NewCart(ctx, cart))
}
