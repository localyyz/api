package checkout

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
	"github.com/pressly/chi/render"
	"github.com/pressly/lg"
	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/web/api"
)

// Create checkout on shopify
func CreateCheckout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)
	cart := ctx.Value("cart").(*data.Cart)

	if cart.Status == data.CartStatusComplete {
		// can't create another checkout on a completed cart
		return
	}

	// Find the product variants
	var variants []*data.ProductVariant
	err := data.DB.Select("pv.*").
		From("cart_items ci").
		LeftJoin("product_variants pv").
		On("ci.variant_id = pv.id").
		Where(db.Cond{"ci.cart_id": cart.ID}).
		All(&variants)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	// group them by places, as line items
	lineItemsMap := map[int64][]*shopify.LineItem{}
	var outOfStock []int64
	for _, v := range variants {
		if v.Limits == 0 {
			// check if any variants are out of stock
			// collect and continue
			outOfStock = append(outOfStock, v.ID)
			continue
		}
		lineItemsMap[v.PlaceID] = append(
			lineItemsMap[v.PlaceID],
			&shopify.LineItem{
				VariantID: v.OfferID,
				Quantity:  1, // TODO: for now, 1 item, hardcoded
			},
		)
	}
	if len(outOfStock) > 0 {
		render.Respond(w, r, api.ErrOutOfStock(outOfStock))
		return
	}

	checkoutMap := make(map[int64]*shopify.Checkout)
	credMap := make(map[int64]*data.ShopifyCred)
	for placeID, lineItems := range lineItemsMap {
		checkoutMap[placeID] = &shopify.Checkout{
			Email:     user.Email,
			LineItems: lineItems,
		}
		cred, err := data.DB.ShopifyCred.FindByPlaceID(placeID)
		if err != nil {
			lg.Warnf("failed to fetch cred with db error: %+v", err)
			continue
		}
		credMap[placeID] = cred
	}

	cart.Etc.ShopifyData = make(map[int64]*data.CartShopifyData)
	for placeID, checkout := range checkoutMap {
		cred := credMap[placeID]

		cl := shopify.NewClient(nil, cred.AccessToken)
		cl.BaseURL, _ = url.Parse(cred.ApiURL)

		// check for line item quantity, return if any is out of stock
		var outOfStock []int64
		for _, v := range variants {
			// TODO: optimize? bulk fetch? is there an api?
			// TODO: return multiple?
			if v.PlaceID != placeID {
				continue
			}

			isStocked, _, _ := cl.Product.IsInStock(ctx, v.OfferID)
			if !isStocked {
				outOfStock = append(outOfStock, v.ID)
			}
			// async mark variant as soldout
			go func(v *data.ProductVariant) {
				v.Limits = 0
				data.DB.ProductVariant.Save(v)
			}(v)
		}
		if len(outOfStock) > 0 {
			render.Respond(w, r, api.ErrOutOfStock(outOfStock))
			return
		}

		cc, _, err := cl.Checkout.Create(ctx, &shopify.CheckoutRequest{checkout})
		if err != nil {
			render.Respond(w, r, err)
			return
		}

		cart.Etc.ShopifyData[placeID] = &data.CartShopifyData{
			Token:            cc.Token,
			CustomerID:       cc.CustomerID,
			Name:             cc.Name,
			PaymentAccountID: cc.PaymentAccountID,
			WebURL:           cc.WebURL,
			WebProcessingURL: cc.WebProcessingURL,
			SubtotalPrice:    atoi(cc.SubtotalPrice),
			TotalPrice:       atoi(cc.TotalPrice),
			TotalTax:         atoi(cc.TotalTax),
			PaymentDue:       cc.PaymentDue,
		}
	}
	cart.Status = data.CartStatusCheckout
	if err := data.DB.Cart.Save(cart); err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Render(w, r, presenter.NewCart(ctx, cart))
}

// update checkout
func UpdateCheckout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cart := ctx.Value("cart").(*data.Cart)

	if cart.Status == data.CartStatusComplete {
		return
	}

	// transform address
	address := &shopify.CustomerAddress{
		Address1:  cart.Etc.ShippingAddress.Address,
		Address2:  cart.Etc.ShippingAddress.AddressOpt,
		City:      cart.Etc.ShippingAddress.City,
		Country:   cart.Etc.ShippingAddress.Country,
		FirstName: cart.Etc.ShippingAddress.FirstName,
		LastName:  cart.Etc.ShippingAddress.LastName,
		Province:  cart.Etc.ShippingAddress.Province,
		Zip:       cart.Etc.ShippingAddress.Zip,
	}

	for placeID, sh := range cart.Etc.ShopifyData {
		cred, _ := data.DB.ShopifyCred.FindByPlaceID(placeID)

		cl := shopify.NewClient(nil, cred.AccessToken)
		cl.BaseURL, _ = url.Parse(cred.ApiURL)
		cl.Debug = true

		checkout := &shopify.Checkout{
			Token:           sh.Token,
			ShippingAddress: address,
			BillingAddress:  address,
		}
		if cart.Etc.ShippingMethods != nil {
			if m, ok := cart.Etc.ShippingMethods[placeID]; ok && m != nil {
				checkout.ShippingLine = &shopify.ShippingLine{
					Handle: m.Handle,
				}
			}
		}
		c, _, err := cl.Checkout.Update(ctx, &shopify.CheckoutRequest{checkout})
		if err != nil {
			lg.Warn(errors.Wrapf(err, "failed to update shopify(%v)", placeID))
			continue
		}

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

func atoi(s string) int64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		lg.Errorf("failed to parse %s to float", s)
		return 0
	}
	return int64(f * 100.0)
}
