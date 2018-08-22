package shopper

import (
	"context"
	"net/url"
	"strconv"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"github.com/pkg/errors"
	db "upper.io/db.v3"
)

type Shopper interface {
	Do(ctx context.Context) error
	HandleError(*shopify.Checkout, error)
	Finalize(req *shopify.Checkout) error
	GetItems() ([]*data.ProductVariant, error)
}

// Checkout implements the Shopper interface
//  as in Shopper performs Checkout
type Checkout struct {
	*data.Checkout
	ShippingAddress *shopify.CustomerAddress
	BillingAddress  *shopify.CustomerAddress
	PaymentCard     *PaymentCard

	Email string
	Err   *CheckoutError
}

var _ interface {
	Shopper
} = &Checkout{}

func (c *Checkout) doCheckout(ctx context.Context, req *shopify.Checkout) (*shopify.Checkout, error) {
	cred, err := data.DB.ShopifyCred.FindOne(db.Cond{"place_id": c.PlaceID})
	if err != nil {
		return req, err
	}

	client := shopify.NewClient(nil, cred.AccessToken)
	client.BaseURL, _ = url.Parse(cred.ApiURL)
	//client.Debug = true

	update, _, err := client.Checkout.CreateOrUpdate(ctx, req)
	if err != nil {
		return req, err
	}
	// validate discount code is applied. if applicable
	if update.AppliedDiscount != nil {
		if reason := update.AppliedDiscount.NonApplicableReason; len(reason) != 0 {
			return req, errors.Wrap(ErrDiscountCode, reason)
		}
	}

	//TODO: make proper shipping. For now, pick the cheapest one.
	rates, _, err := client.Checkout.ListShippingRates(ctx, req.Token)
	if len(rates) == 0 {
		// does not ship to this address
		return req, ErrShippingRates
	}
	if err != nil {
		return req, err
	}

	// update checkout shipping line
	req.ShippingLine = &shopify.ShippingLine{
		Handle: rates[0].Handle,
	}

	//for now, pick the first rate and apply to the checkout.. TODO make this proper
	//sync the cart shipping method

	// NOTE: shopify checkout update cannot update shipping_line and discount
	// code in the same request. Or it will throw a cryptic error message about
	// shipping line invalid
	//
	// remove it.
	discountCode := req.DiscountCode
	req.DiscountCode = ""
	if _, _, err := client.Checkout.Update(ctx, req); err != nil {
		return req, err
	}
	// put it back
	req.DiscountCode = discountCode

	//Cache delivery range. *NOTE: shopify doesn't return the delivery range
	//when updating the shipping line.. need to pull it from the rates
	req.ShippingLine.DeliveryRange = rates[0].DeliveryRange

	return req, nil
}

func NewCheckout(ctx context.Context, checkout *data.Checkout) *Checkout {
	shippingAddress := ctx.Value(ShippingAddressCtxKey).(*data.CartAddress)
	billingAddress := ctx.Value(ShippingAddressCtxKey).(*data.CartAddress)
	email := ctx.Value(EmailCtxKey).(string)
	return &Checkout{
		Checkout:        checkout,
		Email:           email,
		ShippingAddress: toShopifyAddress(shippingAddress),
		BillingAddress:  toShopifyAddress(billingAddress),
	}
}

func (c *Checkout) HandleError(req *shopify.Checkout, err error) {
	// check applied discount
	if e := errors.Cause(err); e == ErrDiscountCode {
		c.Err = &CheckoutError{
			ErrCode: CheckoutErrorCodeDiscountCode,
			Err:     err,
		}
		return
	}
	if err == nil {
		return
	}
	if e, ok := err.(*shopify.LineItemError); ok && e != nil && e.Position != "" {
		idx, _ := strconv.Atoi(e.Position)
		c.Err = &CheckoutError{
			ItemID:  req.LineItems[idx].VariantID,
			ErrCode: CheckoutErrorCodeLineItem,
			Err:     ErrItemOutofStock,
		}
		return
	}
	if e, ok := err.(*shopify.AddressError); ok && e != nil {
		code := CheckoutErrorCodeShippingAddress
		if e.Type() == "billing_address" {
			code = CheckoutErrorCodeBillingAddress
		}
		c.Err = &CheckoutError{
			PlaceID: c.PlaceID,
			ErrCode: code,
			Err:     err,
		}
		return
	}
	if err == ErrShippingRates {
		c.Err = &CheckoutError{
			PlaceID: c.PlaceID,
			ErrCode: CheckoutErrorCodeNoShipping,
			Err:     err,
		}
		return
	}
	c.Err = &CheckoutError{
		PlaceID: c.PlaceID,
		Err:     err,
		ErrCode: CheckoutErrorCodeGeneric,
	}
}

func (c *Checkout) GetItems() ([]*data.ProductVariant, error) {
	// Find the cart item -> product variants
	cartItems, err := data.DB.CartItem.FindByCheckoutID(c.ID)
	if err != nil {
		return nil, err
	}
	variantIDs := make([]int64, len(cartItems))
	for i, ci := range cartItems {
		variantIDs[i] = ci.VariantID
	}
	return data.DB.ProductVariant.FindAll(db.Cond{"id": variantIDs})
}

func (c *Checkout) Finalize(req *shopify.Checkout) error {
	// metadata
	c.Token = req.Token
	c.Name = req.Name
	c.WebURL = req.WebURL
	c.CustomerID = req.CustomerID
	c.Currency = req.Currency
	c.PaymentAccountID = req.ShopifyPaymentAccountID

	// shipping
	if req.ShippingLine != nil {
		// shipping
		c.ShippingLine = &data.CheckoutShippingLine{
			ShippingLine: req.ShippingLine,
		}
		c.TotalShipping, _ = strconv.ParseFloat(req.ShippingLine.Price, 64)
	}

	// tax
	c.TaxesIncluded = req.TaxesIncluded
	c.TotalTax, _ = strconv.ParseFloat(req.TotalTax, 64)
	for _, t := range req.TaxLines {
		c.TaxLines = append(c.TaxLines, &data.CheckoutTaxLine{TaxLine: t})
	}

	// price
	c.SubtotalPrice, _ = strconv.ParseFloat(req.SubtotalPrice, 64)
	c.TotalPrice, _ = strconv.ParseFloat(req.TotalPrice, 64)
	c.PaymentDue = req.PaymentDue

	// discount
	c.AppliedDiscount = &data.CheckoutAppliedDiscount{
		AppliedDiscount: req.AppliedDiscount,
	}
	if c.AppliedDiscount != nil && !c.AppliedDiscount.Applicable {
		// if discount code is non applicable after some change. remove it
		c.DiscountCode = ""
	}

	return data.DB.Checkout.Save(c.Checkout)
}

func (c *Checkout) Do(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	// fetch checkout items
	variants, err := c.GetItems()
	if err != nil || len(variants) == 0 {
		// if nothing to checkout, return error
		if err == nil {
			return ErrEmptyCheckout
		}
		return err
	}

	// create the checkout request
	req := &shopify.Checkout{
		ShippingAddress: c.ShippingAddress,
		BillingAddress:  c.BillingAddress,
		Email:           c.Email,
		Token:           c.Token,
		DiscountCode:    c.DiscountCode,
	}

	for _, v := range variants {
		req.LineItems = append(req.LineItems, &shopify.LineItem{
			VariantID: v.OfferID,
			Quantity:  1, // TODO: for now, 1 item, hardcoded
		})
	}

	c.HandleError(c.doCheckout(ctx, req))
	return c.Finalize(req)
}
