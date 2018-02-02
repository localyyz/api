package express

import (
	"net/http"
	"net/url"
	"strconv"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/render"
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

	// save the cart item
	if err := data.DB.CartItem.Save(&data.CartItem{
		CartID:    cart.ID,
		ProductID: payload.ProductID,
		VariantID: variant.ID,
		PlaceID:   variant.PlaceID,
		Quantity:  uint32(payload.Quantity)}); err != nil {

		render.Respond(w, r, err)
		return
	}

	// find the shopify cred for the merchant and start the checkout process
	creds, err := data.DB.ShopifyCred.FindByPlaceID(variant.PlaceID)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	cl := shopify.NewClient(nil, creds.AccessToken)
	cl.BaseURL, _ = url.Parse(creds.ApiURL)
	checkout, _, err := cl.Checkout.Create(
		ctx,
		&shopify.CheckoutRequest{
			Checkout: &shopify.Checkout{
				// TODO: user email? -> need to update cart with user email
				LineItems: []*shopify.LineItem{{VariantID: variant.OfferID, Quantity: 1}},
			},
		},
	)
	if err != nil || checkout == nil {
		lg.Alertf("failed to create new checkout place(%d) with err: %+v", variant.PlaceID, err)
		render.Respond(w, r, err)
		return
	}
	cart.Etc.ShopifyData[variant.PlaceID] = &data.CartShopifyData{
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
	}
	data.DB.Cart.Save(cart)

	presented := &expressCheckoutResponse{
		Cart: presenter.NewCart(ctx, cart),
	}

	checkoutShippings, _, _ := cl.Checkout.ListShippingRates(ctx, checkout.Token)
	for _, shipping := range checkoutShippings {
		presented.Rates = append(presented.Rates, &data.CartShippingMethod{
			Handle:        shipping.ID,
			Title:         shipping.Title,
			Price:         atoi(shipping.Price),
			DeliveryRange: shipping.DeliveryRange,
		})
	}
	render.Render(w, r, presented)
}

func CreatePayment(w http.ResponseWriter, r *http.Request) {

}

func DeleteCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cart := ctx.Value("cart").(*data.Cart)
	data.DB.Cart.Delete(cart)
}

func atoi(s string) int64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		lg.Errorf("failed to parse %s to float", s)
		return 0
	}
	return int64(f * 100.0)
}
