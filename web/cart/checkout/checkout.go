package checkout

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/render"
	"github.com/pkg/errors"
	"github.com/pressly/lg"
	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/web/api"
)

type cartPayment struct {
	Number string `json:"number"`
	Type   string `json:"type"`
	Expiry string `json:"expiry"`
	Name   string `json:"name"`
	CVC    string `json:"cvc"`
}

type checkoutRequest struct {
	ShippingAddress *data.CartAddress        `json:"shippingAddress,omitempty"`
	BillingAddress  *data.CartAddress        `json:"billingAddress,omitempty"`
	Shipping        *data.CartShippingMethod `json:"shipping"`
	DiscountCode    string                   `json:"discountCode,omitempty"`
}

func (c *checkoutRequest) Bind(r *http.Request) error {
	return nil
}

// Create checkout on shopify
func CreateCheckout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)
	cart := ctx.Value("cart").(*data.Cart)

	if cart.Status == data.CartStatusComplete {
		// can't create another checkout on a completed cart
		return
	}

	var payload checkoutRequest
	if err := render.Bind(r, &payload); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	// Find the cart item -> product variants
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
	// + check variant stock on our side. if any is out-of-stock
	// return with response
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
		render.Respond(w, r, api.ErrOutOfStockCart(outOfStock))
		return
	}

	// transform address to shopify
	shippingAddress := &shopify.CustomerAddress{
		Address1:  payload.ShippingAddress.Address,
		Address2:  payload.ShippingAddress.AddressOpt,
		City:      payload.ShippingAddress.City,
		Country:   payload.ShippingAddress.Country,
		FirstName: payload.ShippingAddress.FirstName,
		LastName:  payload.ShippingAddress.LastName,
		Province:  payload.ShippingAddress.Province,
		Zip:       payload.ShippingAddress.Zip,
	}
	billingAddress := &shopify.CustomerAddress{
		Address1:  payload.BillingAddress.Address,
		Address2:  payload.BillingAddress.AddressOpt,
		City:      payload.BillingAddress.City,
		Country:   payload.BillingAddress.Country,
		FirstName: payload.BillingAddress.FirstName,
		LastName:  payload.BillingAddress.LastName,
		Province:  payload.BillingAddress.Province,
		Zip:       payload.BillingAddress.Zip,
	}

	checkoutMap := make(map[int64]*shopify.Checkout)
	credMap := make(map[int64]*data.ShopifyCred)
	for placeID, lineItems := range lineItemsMap {
		checkoutMap[placeID] = &shopify.Checkout{
			Email:           user.Email,
			LineItems:       lineItems,
			ShippingAddress: shippingAddress,
			BillingAddress:  billingAddress,
		}
		cred, err := data.DB.ShopifyCred.FindByPlaceID(placeID)
		if err != nil {
			lg.Warnf("failed to fetch cred with db error: %+v", err)
			continue
		}
		credMap[placeID] = cred
	}

	// create shipify data holder on the cart
	cart.Etc.ShopifyData = make(map[int64]*data.CartShopifyData)
	cart.Etc.ShippingMethods = make(map[int64]*data.CartShippingMethod)
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

			s, _, err := cl.Product.GetStock(ctx, v.OfferID)
			if err != nil {
				render.Respond(w, r, err)
				return
			}
			if s == 0 {
				outOfStock = append(outOfStock, v.ID)
			}

			if ss := int64(s); v.Limits != ss {
				// update the quantity
				v.Limits = ss
				data.DB.ProductVariant.Save(v)
			}
		}
		if len(outOfStock) > 0 {
			// return right away if something is out of stock
			render.Render(w, r, api.ErrOutOfStockCart(outOfStock))
			return
		}

		cc, _, err := cl.Checkout.Create(ctx, &shopify.CheckoutRequest{checkout})
		if err != nil {
			lg.Alertf("checkout create: cart(%d) pl(%d) user(%d) create failed: %+v", cart.ID, placeID, user.ID, err)
			render.Respond(w, r, err)
			return
		}

		updateCheckout := &shopify.Checkout{
			Token: cc.Token,
		}
		// check the shipping rate
		// TODO: make proper shipping. For now, pick the cheapest one.
		rates, _, _ := cl.Checkout.ListShippingRates(ctx, cc.Token)
		// for now, pick the first rate and apply to the checkout
		// make this proper
		if len(rates) > 0 {
			//update the checkout with the shipping line
			updateCheckout.ShippingLine = &shopify.ShippingLine{Handle: rates[0].Handle}
			cart.Etc.ShippingMethods[placeID] = &data.CartShippingMethod{
				rates[0].Handle,
				rates[0].Title,
				atoi(rates[0].Price),
				rates[0].DeliveryRange,
			}
		}

		cc, _, err = cl.Checkout.Update(
			ctx,
			&shopify.CheckoutRequest{Checkout: updateCheckout},
		)
		if err != nil {
			lg.Alertf("checkout (%d) shipping failed: %+v", err)
		}
		cart.Etc.ShopifyData[placeID] = &data.CartShopifyData{
			Token:      cc.Token,
			CustomerID: cc.CustomerID,
			Name:       cc.Name,
			ShopifyPaymentAccountID: cc.ShopifyPaymentAccountID,
			PaymentURL:              cc.PaymentURL,
			WebURL:                  cc.WebURL,
			WebProcessingURL:        cc.WebProcessingURL,
			TotalTax:                atoi(cc.TotalTax),
			TotalPrice:              atoi(cc.TotalPrice),
			PaymentDue:              cc.PaymentDue,
		}
		if cc.AppliedDiscount != nil && cc.AppliedDiscount.Applicable {
			cart.Etc.ShopifyData[placeID].Discount = cc.AppliedDiscount
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

	var payload checkoutRequest
	if err := render.Bind(r, &payload); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	var checkout shopify.Checkout
	if sh := payload.ShippingAddress; sh != nil {
		checkout.ShippingAddress = &shopify.CustomerAddress{
			Address1:  sh.Address,
			Address2:  sh.AddressOpt,
			City:      sh.City,
			Country:   sh.Country,
			FirstName: sh.FirstName,
			LastName:  sh.LastName,
			Province:  sh.Province,
			Zip:       sh.Zip,
		}
		cart.Etc.ShippingAddress = payload.ShippingAddress
	}
	if bs := payload.BillingAddress; bs != nil {
		checkout.BillingAddress = &shopify.CustomerAddress{
			Address1:  bs.Address,
			Address2:  bs.AddressOpt,
			City:      bs.City,
			Country:   bs.Country,
			FirstName: bs.FirstName,
			LastName:  bs.LastName,
			Province:  bs.Province,
			Zip:       bs.Zip,
		}
		cart.Etc.BillingAddress = payload.BillingAddress
	}

	for placeID, sh := range cart.Etc.ShopifyData {
		cred, _ := data.DB.ShopifyCred.FindByPlaceID(placeID)

		cl := shopify.NewClient(nil, cred.AccessToken)
		cl.BaseURL, _ = url.Parse(cred.ApiURL)
		cl.Debug = true

		ch := checkout

		if payload.DiscountCode != "" {
			ch.DiscountCode = payload.DiscountCode
		}

		ch.Token = sh.Token
		c, _, err := cl.Checkout.Update(ctx, &shopify.CheckoutRequest{&ch})
		if err != nil {
			lg.Warn(errors.Wrapf(err, "failed to update shopify(%v)", placeID))
			continue
		}
		cart.Etc.ShopifyData[placeID].TotalPrice = atoi(c.TotalPrice)
		cart.Etc.ShopifyData[placeID].TotalTax = atoi(c.TotalTax)
		cart.Etc.ShopifyData[placeID].PaymentDue = c.PaymentDue
		cart.Etc.ShopifyData[placeID].Discount = c.AppliedDiscount
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
