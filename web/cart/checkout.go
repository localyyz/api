package cart

import (
	"context"
	"net/http"
	"net/url"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/render"
	"github.com/pkg/errors"
	"github.com/pressly/lg"
	set "gopkg.in/fatih/set.v0"
	db "upper.io/db.v3"
)

func fetchCartItemVariant(cart *data.Cart) ([]*data.ProductVariant, error) {
	// Find the cart item -> product variants
	cartItems, err := data.DB.CartItem.FindByCartID(cart.ID)
	if err != nil {
		return nil, err
	}

	variantIDs := make([]int64, len(cartItems))
	for i, ci := range cartItems {
		variantIDs[i] = ci.VariantID
	}

	return data.DB.ProductVariant.FindAll(db.Cond{"id": variantIDs})
}

func validateCheckout(cart *data.Cart) error {
	if cart.Etc.ShippingAddress == nil {
		return errors.New("empty or invalid shipping address")
	}
	if cart.Etc.BillingAddress == nil {
		return errors.New("empty or invalid billing address")
	}
	return nil
}

func validateVariants(variants []*data.ProductVariant) error {
	var outOfStock []int64
	for _, v := range variants {
		if v.Limits == 0 {
			// check if any variants are out of stock
			// collect and continue
			outOfStock = append(outOfStock, v.OfferID)
			continue
		}
	}
	if len(outOfStock) > 0 {
		return api.ErrOutOfStockCart(outOfStock)
	}
	return nil
}

func createCheckout(ctx context.Context, cl *shopify.Client, checkout *shopify.Checkout) error {
	// create the checkout on shopify
	cc, _, err := cl.Checkout.Create(ctx, &shopify.CheckoutRequest{checkout})
	if err != nil {
	}
	// end create checkout

	// TODO: make proper shipping. For now, pick the cheapest one.
	rates, _, err := cl.Checkout.ListShippingRates(ctx, cc.Token)
	if err != nil {
		return err
	}
	if len(rates) == 0 {
		return errors.New("invalid shipping address")
	}
	// for now, pick the first rate and apply to the checkout.. TODO make this proper
	// sync the cart shipping method
	cc, _, err = cl.Checkout.Update(
		ctx,
		&shopify.CheckoutRequest{
			Checkout: &shopify.Checkout{
				Token: cc.Token,
				ShippingLine: &shopify.ShippingLine{
					Handle: rates[0].Handle,
				},
			},
		},
	)
	if err != nil {
		return err
	}

	checkout = cc
	checkout.ShippingLine.DeliveryRange = rates[0].DeliveryRange

	return nil
}

// Create checkout on shopify
func CreateCheckout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)
	cart := ctx.Value("cart").(*data.Cart)

	if cart.Status != data.CartStatusInProgress {
		// can't create another checkout on a completed cart
		err := errors.New("invalid cart status, already completed.")
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	// TODO verify cart checkout fields
	if err := validateCheckout(cart); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	variants, err := fetchCartItemVariant(cart)
	if err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	if err := validateVariants(variants); err != nil {
		render.Respond(w, r, err)
		return
	}

	merchantIDset := set.New()
	for _, v := range variants {
		merchantIDset.Add(v.PlaceID)
	}
	creds, err := data.DB.ShopifyCred.FindAll(
		db.Cond{
			"place_id": set.IntSlice(merchantIDset),
		},
	)
	if err != nil {
		render.Respond(w, r, api.ErrInvalidRequest(err))
		return
	}

	// group them by places, as line items
	// + check variant stock on our side. if any is out-of-stock
	// return with response
	lineItemsMap := map[int64][]*shopify.LineItem{}
	for _, v := range variants {
		lineItemsMap[v.PlaceID] = append(
			lineItemsMap[v.PlaceID],
			&shopify.LineItem{
				VariantID: v.OfferID,
				Quantity:  1, // TODO: for now, 1 item, hardcoded
			},
		)
	}

	//transform address to shopify
	shippingAddress := &shopify.CustomerAddress{
		Address1:  cart.Etc.ShippingAddress.Address,
		Address2:  cart.Etc.ShippingAddress.AddressOpt,
		City:      cart.Etc.ShippingAddress.City,
		Country:   cart.Etc.ShippingAddress.Country,
		FirstName: cart.Etc.ShippingAddress.FirstName,
		LastName:  cart.Etc.ShippingAddress.LastName,
		Province:  cart.Etc.ShippingAddress.Province,
		Zip:       cart.Etc.ShippingAddress.Zip,
	}
	billingAddress := &shopify.CustomerAddress{
		Address1:  cart.Etc.BillingAddress.Address,
		Address2:  cart.Etc.BillingAddress.AddressOpt,
		City:      cart.Etc.BillingAddress.City,
		Country:   cart.Etc.BillingAddress.Country,
		FirstName: cart.Etc.BillingAddress.FirstName,
		LastName:  cart.Etc.BillingAddress.LastName,
		Province:  cart.Etc.BillingAddress.Province,
		Zip:       cart.Etc.BillingAddress.Zip,
	}

	// create shipify data holder on the cart
	cart.Etc.ShopifyData = make(map[int64]*data.CartShopifyData)
	cart.Etc.ShippingMethods = make(map[int64]*data.CartShippingMethod)
	for _, cred := range creds {
		cl := shopify.NewClient(nil, cred.AccessToken)
		cl.BaseURL, _ = url.Parse(cred.ApiURL)

		checkout := &shopify.Checkout{
			Email:           user.Email,
			LineItems:       lineItemsMap[cred.PlaceID],
			ShippingAddress: shippingAddress,
			BillingAddress:  billingAddress,
		}
		if err := createCheckout(ctx, cl, checkout); err != nil {
			err := errors.Wrapf(err, "checkout: cart(%d) pl(%d) user(%d)", cart.ID, cred.PlaceID, user.ID, err)
			lg.Alert(err)

			// TODO: what do we do here? how do we tell frontend we completely borked
			// this update?
			continue
		}

		cart.Etc.ShippingMethods[cred.PlaceID] = &data.CartShippingMethod{
			Handle:        checkout.ShippingLine.Handle,
			Title:         checkout.ShippingLine.Title,
			Price:         atoi(checkout.ShippingLine.Price),
			DeliveryRange: checkout.ShippingLine.DeliveryRange,
		}
		cart.Etc.ShopifyData[cred.PlaceID] = &data.CartShopifyData{
			Token:      checkout.Token,
			CustomerID: checkout.CustomerID,
			Name:       checkout.Name,
			ShopifyPaymentAccountID: checkout.ShopifyPaymentAccountID,
			PaymentURL:              checkout.PaymentURL,
			WebURL:                  checkout.WebURL,
			WebProcessingURL:        checkout.WebProcessingURL,

			TotalTax:   atoi(checkout.TotalTax),
			TotalPrice: atoi(checkout.TotalPrice),
			PaymentDue: checkout.PaymentDue,
		}
		if checkout.AppliedDiscount != nil && checkout.AppliedDiscount.Applicable {
			cart.Etc.ShopifyData[cred.PlaceID].Discount = checkout.AppliedDiscount
		}
	}

	cart.Status = data.CartStatusCheckout
	if err := data.DB.Cart.Save(cart); err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Render(w, r, presenter.NewCart(ctx, cart))
}
