package cart

import (
	"context"
	"encoding/json"
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

var (
	ErrShippingRates   = errors.New("merchant does not ship to your address")
	ErrEmptyCheckout   = errors.New("nothing in your cart")
	ErrInvalidShipping = errors.New("empty or invalid shipping address")
	ErrInvalidBilling  = errors.New("empty or invalid billing address")
	ErrInvalidStatus   = errors.New("invalid cart status, already completed.")
	ErrItemOutofStock  = errors.New("item is out of stock")
)

type checkoutError struct {
	placeID int64
	itemID  int64
	errCode checkoutErrorCode
	err     error
}

type checkoutEmail struct {
	Email string
}

type checkoutErrorCode uint32

const (
	_ checkoutErrorCode = iota
	CheckoutErrorCodeGeneric
	CheckoutErrorCodeLineItem
	CheckoutErrorCodeNoShipping
	CheckoutErrorCodeShippingAddress
)

func (e *checkoutError) Error() string {
	return e.err.Error()
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
		return ErrInvalidShipping
	}
	if cart.Etc.BillingAddress == nil {
		return ErrInvalidBilling
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
	if err != nil || len(rates) == 0 {
		if err == nil {
			// does not ship to this address
			err = ErrShippingRates
		}
		return err
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

type universalCheckout struct {
	shippingAddress *shopify.CustomerAddress
	billingAddress  *shopify.CustomerAddress

	lineItemsMap map[int64][]*shopify.LineItem
	cart         *data.Cart
}

func (u *universalCheckout) createMerchantCheckout(ctx context.Context, placeID int64, client *shopify.Client, userEmail string) error {

	cart := ctx.Value("cart").(*data.Cart)

	checkout := &shopify.Checkout{
		Email:           userEmail,
		LineItems:       u.lineItemsMap[placeID],
		ShippingAddress: u.shippingAddress,
		BillingAddress:  u.billingAddress,
	}
	// check if there is an existing checkout, if there is, set the checkout token
	if sh, ok := cart.Etc.ShopifyData[placeID]; sh != nil && ok {
		checkout.Token = sh.Token
	}

	// TODO: save checkout token even on error. rignt now
	// we're creating multiple checkouts if remote fails

	// find and apply discount code that's from this merchant
	if discount, _ := data.DB.PlaceDiscount.FindByPlaceID(placeID); discount != nil {
		checkout.DiscountCode = discount.Code
	}

	if err := createCheckout(ctx, client, checkout); err != nil {
		if le, ok := err.(*shopify.LineItemError); ok && le != nil && le.Position != "" {
			idx, _ := strconv.Atoi(le.Position)
			return &checkoutError{
				itemID:  u.lineItemsMap[placeID][idx].VariantID,
				errCode: CheckoutErrorCodeLineItem,
				err:     ErrItemOutofStock,
			}
		}
		if le, ok := err.(*shopify.ShippingAddressError); ok && le != nil {
			return &checkoutError{
				placeID: placeID,
				errCode: CheckoutErrorCodeShippingAddress,
				err:     err,
			}
		}
		if err == ErrShippingRates {
			return &checkoutError{
				placeID: placeID,
				errCode: CheckoutErrorCodeNoShipping,
				err:     err,
			}
		}
		return &checkoutError{
			placeID: placeID,
			err:     err,
			errCode: CheckoutErrorCodeGeneric,
		}
	}

	cart.Etc.ShippingMethods[placeID] = &data.CartShippingMethod{
		Handle:        checkout.ShippingLine.Handle,
		Title:         checkout.ShippingLine.Title,
		Price:         atoi(checkout.ShippingLine.Price),
		DeliveryRange: checkout.ShippingLine.DeliveryRange,
	}
	cart.Etc.ShopifyData[placeID] = &data.CartShopifyData{
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
		cart.Etc.ShopifyData[placeID].Discount = checkout.AppliedDiscount
	}

	return nil
}

// Create checkout is the entry point to getting a full "checkout"
// object back from the API
func CreateCheckout(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	var email checkoutEmail
	err := decoder.Decode(&email)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	ctx := r.Context()
	cart := ctx.Value("cart").(*data.Cart)

	if cart.Status > data.CartStatusCheckout {
		render.Render(w, r, api.ErrInvalidRequest(ErrInvalidStatus))
		return
	}

	// TODO verify cart checkout fields
	if err := validateCheckout(cart); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	variants, err := fetchCartItemVariant(cart)
	if err != nil || len(variants) == 0 {
		// if nothing to checkout, return error
		if err == nil {
			err = ErrEmptyCheckout
		}
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	merchantIDset := set.New()
	for _, v := range variants {
		merchantIDset.Add(v.PlaceID)
	}
	creds, err := data.DB.ShopifyCred.FindAll(db.Cond{"place_id": int64Slice(merchantIDset)})
	if err != nil {
		render.Respond(w, r, api.ErrInvalidRequest(err))
		return
	}

	u := &universalCheckout{
		// transform address to shopify
		shippingAddress: &shopify.CustomerAddress{
			Address1:  cart.Etc.ShippingAddress.Address,
			Address2:  cart.Etc.ShippingAddress.AddressOpt,
			City:      cart.Etc.ShippingAddress.City,
			Country:   cart.Etc.ShippingAddress.Country,
			FirstName: cart.Etc.ShippingAddress.FirstName,
			LastName:  cart.Etc.ShippingAddress.LastName,
			Province:  cart.Etc.ShippingAddress.Province,
			Zip:       cart.Etc.ShippingAddress.Zip,
		},
		billingAddress: &shopify.CustomerAddress{
			Address1:  cart.Etc.BillingAddress.Address,
			Address2:  cart.Etc.BillingAddress.AddressOpt,
			City:      cart.Etc.BillingAddress.City,
			Country:   cart.Etc.BillingAddress.Country,
			FirstName: cart.Etc.BillingAddress.FirstName,
			LastName:  cart.Etc.BillingAddress.LastName,
			Province:  cart.Etc.BillingAddress.Province,
			Zip:       cart.Etc.BillingAddress.Zip,
		},
	}

	// group them by places, as line items
	u.lineItemsMap = map[int64][]*shopify.LineItem{}
	for _, v := range variants {
		u.lineItemsMap[v.PlaceID] = append(
			u.lineItemsMap[v.PlaceID],
			&shopify.LineItem{
				VariantID: v.OfferID,
				Quantity:  1, // TODO: for now, 1 item, hardcoded
			},
		)
	}

	// create shipify data holder on the cart
	if cart.Etc.ShopifyData == nil {
		cart.Etc.ShopifyData = make(map[int64]*data.CartShopifyData)
	}
	if cart.Etc.ShippingMethods == nil {
		cart.Etc.ShippingMethods = make(map[int64]*data.CartShippingMethod)
	}

	var lastError error
	for _, cred := range creds {
		cl := shopify.NewClient(nil, cred.AccessToken)
		cl.BaseURL, _ = url.Parse(cred.ApiURL)
		cl.Debug = true
		if err := u.createMerchantCheckout(ctx, cred.PlaceID, cl, email.Email); err != nil {
			lg.Alert(errors.Wrapf(err, "checkout: cart(%d) pl(%d)", cart.ID, cred.PlaceID))
			lastError = err
		}
	}

	cart.Status = data.CartStatusCheckout
	if err := data.DB.Cart.Save(cart); err != nil {
		render.Respond(w, r, err)
		return
	}

	presented := presenter.NewCart(ctx, cart)
	if e, ok := lastError.(*checkoutError); e != nil && ok {
		presented.HasError = true
		presented.Error = e.err.Error()
		presented.ErrorCode = uint32(e.errCode)

		switch e.errCode {
		case CheckoutErrorCodeNoShipping, CheckoutErrorCodeShippingAddress:
			presented.ShippingAddress.HasError = true
			presented.ShippingAddress.Error = e.err
		}
		for _, ci := range presented.CartItems {
			ci.HasError = ci.Variant.OfferID == e.itemID
			ci.Error = e.err
		}
	}

	render.Render(w, r, presented)
}
