package cart

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

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

type checkoutError struct {
	placeID int64
	itemID  int64
	err     error
}

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
	var (
		err error
		ch  *shopify.Checkout
	)
	// check if we're updating an already exiting checkout
	if len(checkout.Token) == 0 {
		ch, _, err = cl.Checkout.Create(ctx, &shopify.CheckoutRequest{checkout})
	} else {
		ch, _, err = cl.Checkout.Update(ctx, &shopify.CheckoutRequest{checkout})
	}
	if err != nil {
		return err
	}

	// TODO: make proper shipping. For now, pick the cheapest one.
	rates, _, err := cl.Checkout.ListShippingRates(ctx, ch.Token)
	if err != nil {
		return err
	}
	if len(rates) == 0 {
		return errors.New("invalid shipping address")
	}
	// for now, pick the first rate and apply to the checkout.. TODO make this proper
	// sync the cart shipping method
	ch, _, err = cl.Checkout.Update(
		ctx,
		&shopify.CheckoutRequest{
			Checkout: &shopify.Checkout{
				Token: ch.Token,
				ShippingLine: &shopify.ShippingLine{
					Handle: rates[0].Handle,
				},
			},
		},
	)
	if err != nil {
		return err
	}

	// Cache delivery range. *NOTE: shopify doesn't return the delivery range
	// when updating the shipping line.. need to pull it from the rates
	*checkout = *ch
	checkout.ShippingLine.DeliveryRange = rates[0].DeliveryRange

	return nil
}

// IntSlice is a helper function that returns a slice of ints of s. If
// the set contains mixed types of items only items of type int are returned.
func int64Slice(s set.Interface) []int64 {
	slice := make([]int64, 0)
	for _, item := range s.List() {
		v, ok := item.(int64)
		if !ok {
			continue
		}

		slice = append(slice, v)
	}
	return slice
}

// Create checkout on shopify
func CreateCheckout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)
	cart := ctx.Value("cart").(*data.Cart)

	if cart.Status != data.CartStatusInProgress {
		// can't create another checkout on an already checkout cart
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
			"place_id": int64Slice(merchantIDset),
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

	// transform address to shopify
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
	if cart.Etc.ShopifyData == nil {
		cart.Etc.ShopifyData = make(map[int64]*data.CartShopifyData)
	}
	if cart.Etc.ShippingMethods == nil {
		cart.Etc.ShippingMethods = make(map[int64]*data.CartShippingMethod)
	}

	var lastError *checkoutError
	for _, cred := range creds {
		cl := shopify.NewClient(nil, cred.AccessToken)
		cl.BaseURL, _ = url.Parse(cred.ApiURL)
		cl.Debug = true

		checkout := &shopify.Checkout{
			Email:           user.Email,
			LineItems:       lineItemsMap[cred.PlaceID],
			ShippingAddress: shippingAddress,
			BillingAddress:  billingAddress,
		}
		// check if there is an existing checkout, if there is, set the checkout token
		if sh, ok := cart.Etc.ShopifyData[cred.PlaceID]; sh != nil && ok {
			checkout.Token = sh.Token
		}

		if err := createCheckout(ctx, cl, checkout); err != nil {
			lg.Alert(errors.Wrapf(err, "checkout: cart(%d) pl(%d) user(%d)", cart.ID, cred.PlaceID, user.ID))
			if le, ok := err.(*shopify.LineItemError); ok && le != nil && le.Position != "" {
				idx, _ := strconv.Atoi(le.Position)
				lastError = &checkoutError{
					itemID: lineItemsMap[cred.PlaceID][idx].VariantID,
					err:    errors.New("item is out of stock"),
				}
				continue
			}
			lastError = &checkoutError{
				placeID: cred.PlaceID,
				err:     err,
			}
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

	presented := presenter.NewCart(ctx, cart)
	if lastError != nil {
		presented.Error = lastError.err.Error()
		for _, ci := range presented.CartItems {
			ci.HasError = (ci.PlaceID == lastError.placeID) || (ci.Variant.OfferID == lastError.itemID)
		}
	}

	render.Render(w, r, presented)
}