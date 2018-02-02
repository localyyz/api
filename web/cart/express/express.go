package express

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
)

type cartItemRequest struct {
	ProductID int64  `json:"productId"`
	Color     string `json:"color"`
	Size      string `json:"size"`
	Quantity  uint32 `json:"quantity"`
}

func (*cartItemRequest) Bind(r *http.Request) error {
	return nil
}

type expressCheckoutResponse struct {
	*presenter.Cart
	Rates []*data.CartShippingMethod `json:"rates"`
}

func GetCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cart := ctx.Value("cart").(*data.Cart)
	render.Render(w, r, presenter.NewCart(ctx, cart))
}

type shippingRateRequest struct {
	FirstName   string `json:"firstName,omitempty"`
	LastName    string `json:"lastName,omitempty"`
	Address     string `json:"address,omitempty"`
	City        string `json:"city"`
	Province    string `json:"province"`
	Country     string `json:"country,omitempty"`
	CountryCode string `json:"countryCode"`
	Zip         string `json:"zip,omitempty"`
}

func (p *shippingRateRequest) Bind(r *http.Request) error {
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

func GetShippingRates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cart := ctx.Value("cart").(*data.Cart)

	var payload shippingRateRequest
	if err := render.Bind(r, &payload); err != nil {
		render.Respond(w, r, err)
		return
	}

	var (
		placeID  int64
		checkout *data.CartShopifyData
	)
	for pID, ch := range cart.Etc.ShopifyData {
		placeID = pID
		checkout = ch
		break
	}

	// fetch shipping rate
	// find the shopify cred for the merchant and start the checkout process
	cred, err := data.DB.ShopifyCred.FindByPlaceID(placeID)
	if err != nil {
		render.Respond(w, r, err)
		return
	}
	cl := shopify.NewClient(nil, cred.AccessToken)
	cl.BaseURL, _ = url.Parse(cred.ApiURL)

	_, _, err = cl.Checkout.Update(
		ctx,
		&shopify.CheckoutRequest{
			Checkout: &shopify.Checkout{
				// Partial customer address from apple pay. enough to get shipping rates
				Token: checkout.Token,
				ShippingAddress: &shopify.CustomerAddress{
					FirstName:   payload.FirstName,
					LastName:    payload.LastName,
					Country:     payload.Country,
					CountryCode: payload.CountryCode,
					Province:    payload.Province,
					City:        payload.City,
					Address1:    payload.Address,
					Zip:         payload.Zip,
				},
			},
		},
	)
	if err != nil {
		render.Respond(w, r, errors.Wrap(err, "express cart shipping rate"))
		return
	}

	shopifyRates, _, _ := cl.Checkout.ListShippingRates(ctx, checkout.Token)
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

func CreateCartItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cart := ctx.Value("cart").(*data.Cart)

	var payload cartItemRequest
	if err := render.Bind(r, &payload); err != nil {
		lg.Warn(err)
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	// fetch the variant from given payload (product id, color and size)
	variant, err := data.DB.ProductVariant.FindOne(
		db.Cond{
			"product_id":                   payload.ProductID,
			"limits >=":                    1,
			db.Raw("lower(etc->>'color')"): payload.Color,
			db.Raw("lower(etc->>'size')"):  payload.Size,
		},
	)
	if err != nil {
		if err == db.ErrNoMoreRows {
			render.Render(w, r, api.ErrOutOfStockAdd(err))
			return
		}
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	toSave := &data.CartItem{
		CartID:    cart.ID,
		ProductID: payload.ProductID,
		VariantID: variant.ID,
		PlaceID:   variant.PlaceID,
		Quantity:  uint32(payload.Quantity)}

	// find the shopify cred for the merchant and start the checkout process
	cred, err := data.DB.ShopifyCred.FindByPlaceID(variant.PlaceID)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	cl := shopify.NewClient(nil, cred.AccessToken)
	cl.BaseURL, _ = url.Parse(cred.ApiURL)
	checkout, _, err := cl.Checkout.Create(
		ctx,
		&shopify.CheckoutRequest{
			Checkout: &shopify.Checkout{
				LineItems: []*shopify.LineItem{{VariantID: variant.OfferID, Quantity: 1}},
			},
		},
	)
	if err != nil || checkout == nil {
		lg.Alertf("failed to create new checkout place(%d) with err: %+v", variant.PlaceID, err)
		render.Respond(w, r, err)
		return
	}
	cart.Etc = data.CartEtc{
		ShopifyData: map[int64]*data.CartShopifyData{
			variant.PlaceID: &data.CartShopifyData{
				Token:      checkout.Token,
				CustomerID: checkout.CustomerID,
				Name:       checkout.Name,
				ShopifyPaymentAccountID: checkout.ShopifyPaymentAccountID,
				PaymentURL:              checkout.PaymentURL,
				WebURL:                  checkout.WebURL,
				WebProcessingURL:        checkout.WebProcessingURL,
				TotalTax:                atoi(checkout.TotalTax),
				TotalPrice:              atoi(checkout.TotalPrice),
				PaymentDue:              checkout.PaymentDue,
				Discount:                checkout.AppliedDiscount,
			},
		},
	}
	cart.Status = data.CartStatusCheckout
	data.DB.Cart.Save(cart)

	// save the cart item
	if err := data.DB.CartItem.Save(toSave); err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Render(w, r, presenter.NewCartItem(ctx, toSave))
}

type cartPaymentRequest struct {
	ShippingAddress     *data.CartAddress        `json:"shippingAddress"`
	BillingAddress      *data.CartAddress        `json:"billingAddress"`
	ShippingMethod      *data.CartShippingMethod `json:"shippingMethod"`
	ExpressPaymentToken string                   `json:"expressPaymentToken"`
	Email               string                   `json:"email,omitempty"`
}

func (p *cartPaymentRequest) Bind(r *http.Request) error {
	if p.ShippingAddress == nil {
		return errors.New("invalid shipping address")
	}
	if p.BillingAddress == nil {
		return errors.New("invalid billing address")
	}
	if p.ShippingMethod == nil {
		return errors.New("invalid shipping method")
	}
	if len(p.ExpressPaymentToken) == 0 {
		return errors.New("invalid token")
	}
	if len(p.Email) == 0 {
		user := r.Context().Value("session.user").(*data.User)
		p.Email = user.Email
	}
	return nil
}

func CreatePayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cart := ctx.Value("cart").(*data.Cart)

	lg.Infof("express cart(%d) payment", cart.ID)
	var payload cartPaymentRequest
	if err := render.Bind(r, &payload); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	var (
		placeID     int64
		shopifyData *data.CartShopifyData
	)
	// for express cart, there's only 1 item for 1 merchant
	for pID, sh := range cart.Etc.ShopifyData {
		placeID = pID
		shopifyData = sh
	}

	// shopify throws error if we update shipping address and
	// shipping line in the same request.
	creds, err := data.DB.ShopifyCred.FindByPlaceID(placeID)
	if err != nil {
		render.Respond(w, r, err)
		return
	}
	cl := shopify.NewClient(nil, creds.AccessToken)
	cl.BaseURL, _ = url.Parse(creds.ApiURL)
	cl.Debug = true

	cart.Etc.ShippingAddress = payload.ShippingAddress
	lg.Infof("express cart(%d) payment update shipping addy(%+v) and email", cart.ID, payload.ShippingAddress)
	_, _, err = cl.Checkout.Update(
		ctx,
		&shopify.CheckoutRequest{
			&shopify.Checkout{
				Token: shopifyData.Token,
				ShippingAddress: &shopify.CustomerAddress{
					Address1:  payload.ShippingAddress.Address,
					Address2:  payload.ShippingAddress.AddressOpt,
					City:      payload.ShippingAddress.City,
					Country:   payload.ShippingAddress.Country,
					FirstName: payload.ShippingAddress.FirstName,
					LastName:  payload.ShippingAddress.LastName,
					Province:  payload.ShippingAddress.Province,
					Zip:       payload.ShippingAddress.Zip,
				},
				BillingAddress: &shopify.CustomerAddress{
					Address1:  payload.BillingAddress.Address,
					Address2:  payload.BillingAddress.AddressOpt,
					City:      payload.BillingAddress.City,
					Country:   payload.BillingAddress.Country,
					FirstName: payload.BillingAddress.FirstName,
					LastName:  payload.BillingAddress.LastName,
					Province:  payload.BillingAddress.Province,
					Zip:       payload.BillingAddress.Zip,
				},
				Email: payload.Email,
			},
		},
	)
	if err != nil {
		lg.Alert(errors.Wrapf(err, "failed to pay cart(%d). shopify(%v)", cart.ID, placeID))
		return
	}

	lg.Infof("express cart(%d) payment update shipping method(%+v)", cart.ID, payload.ShippingMethod)
	cc, _, err := cl.Checkout.Update(
		ctx,
		&shopify.CheckoutRequest{
			&shopify.Checkout{
				Token: shopifyData.Token,
				ShippingLine: &shopify.ShippingLine{
					Handle: payload.ShippingMethod.Handle,
				},
			},
		},
	)
	if err != nil {
		lg.Alert(errors.Wrapf(err, "failed to pay cart(%d). shopify(%v)", cart.ID, placeID))
		return
	}

	u, _ := uuid.NewUUID()
	payment := &shopify.PaymentRequest{
		Payment: &shopify.Payment{
			Amount:      cc.CheckoutPrice.PaymentDue,
			UniqueToken: u.String(),
			PaymentToken: &shopify.PaymentToken{
				PaymentData: payload.ExpressPaymentToken,
				Type:        shopify.StripeVaultToken,
			},
			RequestDetails: &shopify.RequestDetail{
				IPAddress: r.RemoteAddr,
			},
		},
	}

	// 4. send payment to shopify
	p, _, err := cl.Checkout.Payment(ctx, shopifyData.Token, payment)
	if err != nil {
		lg.Alertf("payment fail: cart(%d) place(%d) with err %+v", cart.ID, placeID, err)
		render.Respond(w, r, err)
		// TODO: do we return here?
		return
	}

	// check payment transaction
	if p.Transaction == nil {
		// something failed. Try again
		lg.Alertf("cart(%d) failed with empty transaction response", cart.ID)
		render.Respond(w, r, errors.New("payment failed, please try again"))
		return
	}

	if p.Transaction.Status != shopify.TransactionStatusSuccess {
		// something failed. Try again
		lg.Alertf("cart(%d) failed with transaction status %v", cart.ID, p.Transaction.Status)
		render.Respond(w, r, api.ErrCardVaultProcess(fmt.Errorf(p.Transaction.Message)))
		return
	}

	lg.Alertf("express cart(%d) was just paid!", cart.ID)

	// 5. save shopify payment id
	shopifyData.PaymentID = p.ID
	shopifyData.PaymentDue = cc.CheckoutPrice.PaymentDue
	shopifyData.TotalTax = atoi(cc.CheckoutPrice.TotalTax)
	shopifyData.TotalPrice = atoi(cc.CheckoutPrice.TotalPrice)

	// mark checkout as has payed
	cart.Status = data.CartStatusPaymentSuccess
	if err := data.DB.Cart.Save(cart); err != nil {
		lg.Alertf("express cart (%d) payment save failed with %+v", cart.ID, err)
	}
	// TODO: create a customer on stripe after the first
	// tokenization so we can send stripe customer id moving forward

	render.Render(w, r, presenter.NewCart(ctx, cart))
}

func DeleteCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cart := ctx.Value("cart").(*data.Cart)

	if err := data.DB.Cart.Delete(cart); err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Status(r, http.StatusNoContent)
	render.Respond(w, r, "")
}

func atoi(s string) int64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		lg.Errorf("failed to parse %s to float", s)
		return 0
	}
	return int64(f * 100.0)
}
