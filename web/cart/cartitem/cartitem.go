package cartitem

import (
	"net/http"
	"net/url"
	"strconv"

	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
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

func GetCartItem(w http.ResponseWriter, r *http.Request) {
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

	// look up variant
	variant, err := data.DB.ProductVariant.FindOne(
		db.Cond{
			"product_id":                   payload.ProductID,
			"limits >=":                    1,
			db.Raw("lower(etc->>'color')"): payload.Color,
			db.Raw("lower(etc->>'size')"):  payload.Size,
		},
	)
	if err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	newItem := &data.CartItem{
		CartID:    cart.ID,
		ProductID: payload.ProductID,
		VariantID: variant.ID,
		PlaceID:   variant.PlaceID,
		Quantity:  uint32(payload.Quantity),
	}

	if err := data.DB.CartItem.Save(newItem); err != nil {
		render.Respond(w, r, err)
		return
	}

	// if cart is not checked out, let checkout cart do the
	// syncing to shopify
	if cart.Status != data.CartStatusCheckout {
		render.Render(w, r, presenter.NewCartItem(ctx, newItem))
		return
	}
	// check if cart etc shopify data is set
	if cart.Etc.ShopifyData == nil {
		cart.Etc.ShopifyData = make(map[int64]*data.CartShopifyData)
	}

	// add to shopify checkout if needed
	// find the shopify cred for the new item
	creds, err := data.DB.ShopifyCred.FindByPlaceID(newItem.PlaceID)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	cl := shopify.NewClient(nil, creds.AccessToken)
	cl.BaseURL, _ = url.Parse(creds.ApiURL)

	// find the line item
	var lineItems []*shopify.LineItem
	var variants []*data.ProductVariant
	data.DB.Select("pv.*").
		From("cart_items ci").
		LeftJoin("product_variants pv").
		On("ci.variant_id = pv.id").
		Where("ci.place_id = ?", newItem.PlaceID).
		All(&variants)
	for _, v := range variants {
		lineItems = append(lineItems, &shopify.LineItem{
			VariantID: v.OfferID,
			Quantity:  1,
		})
	}

	var token string
	if sh, ok := cart.Etc.ShopifyData[newItem.PlaceID]; ok {
		token = sh.Token
	}

	var cc *shopify.Checkout
	if len(token) != 0 {
		cc, _, _ = cl.Checkout.Update(
			ctx,
			&shopify.CheckoutRequest{
				Checkout: &shopify.Checkout{
					Token:     token,
					LineItems: lineItems,
				},
			},
		)
		cart.Etc.ShopifyData[newItem.PlaceID].TotalTax = atoi(cc.TotalTax)
		cart.Etc.ShopifyData[newItem.PlaceID].TotalPrice = atoi(cc.TotalPrice)
		cart.Etc.ShopifyData[newItem.PlaceID].PaymentDue = cc.PaymentDue
		cart.Etc.ShopifyData[newItem.PlaceID].Discount = cc.AppliedDiscount
	} else {
		cc, _, _ = cl.Checkout.Create(
			ctx,
			&shopify.CheckoutRequest{
				Checkout: &shopify.Checkout{
					Token:     token,
					LineItems: lineItems,
				},
			},
		)
		cart.Etc.ShopifyData[newItem.PlaceID] = &data.CartShopifyData{
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
			Discount:                cc.AppliedDiscount,
		}
	}
	data.DB.Cart.Save(cart)

	render.Render(w, r, presenter.NewCartItem(ctx, newItem))
}

// TODO: add cart items even after checkout is done
func UpdateCartItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cartItem := ctx.Value("cart_item").(*data.CartItem)

	payload := cartItemRequest{}
	if err := render.Bind(r, &payload); err != nil {
		lg.Warn(err)
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	// mostly just quantity
	cartItem.Quantity = payload.Quantity
	if err := data.DB.Save(cartItem); err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Render(w, r, presenter.NewCartItem(ctx, cartItem))
}

func RemoveCartItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cartItem := ctx.Value("cart_item").(*data.CartItem)
	cart := ctx.Value("cart").(*data.Cart)
	if err := data.DB.CartItem.Delete(cartItem); err != nil {
		render.Respond(w, r, err)
		return
	}

	// if the cart hasn't been checked out yet. no need
	// to sync status to the stores
	if cart.Status != data.CartStatusCheckout {
		render.Status(r, http.StatusNoContent)
		render.Respond(w, r, "")
		return
	}

	// 1. Check if this cart has been "checked" out, and
	// that this specific store has a checkout created
	var token string
	if sh, ok := cart.Etc.ShopifyData[cartItem.PlaceID]; ok {
		token = sh.Token
	} else {
		// if this store was never checked out, no need
		// to sync to shopify
		render.Status(r, http.StatusNoContent)
		render.Respond(w, r, "")
		return
	}

	var lineItems []*shopify.LineItem
	var variants []*data.ProductVariant
	data.DB.Select("pv.*").
		From("cart_items ci").
		LeftJoin("product_variants pv").
		On("ci.variant_id = pv.id").
		Where("ci.place_id = ?", cartItem.PlaceID).
		All(&variants)
	for _, v := range variants {
		if v.ID == cartItem.VariantID {
			continue
		}
		lineItems = append(lineItems, &shopify.LineItem{
			VariantID: v.OfferID,
			Quantity:  1,
		})
	}

	if len(lineItems) == 0 {
		// if there are no line items. remove the checkout from shopify data
		delete(cart.Etc.ShopifyData, cartItem.PlaceID)
		// mark cart status as in progress
		cart.Status = data.CartStatusInProgress
	} else {
		// push the change to shopify
		creds, err := data.DB.ShopifyCred.FindByPlaceID(cartItem.PlaceID)
		if err != nil {
			render.Respond(w, r, err)
			return
		}

		cl := shopify.NewClient(nil, creds.AccessToken)
		cl.BaseURL, _ = url.Parse(creds.ApiURL)

		cc, _, _ := cl.Checkout.Update(
			ctx,
			&shopify.CheckoutRequest{
				Checkout: &shopify.Checkout{
					Token:     token,
					LineItems: lineItems,
				},
			},
		)
		cart.Etc.ShopifyData[cartItem.PlaceID].TotalTax = atoi(cc.TotalTax)
		cart.Etc.ShopifyData[cartItem.PlaceID].TotalPrice = atoi(cc.TotalPrice)
		cart.Etc.ShopifyData[cartItem.PlaceID].PaymentDue = cc.PaymentDue
		cart.Etc.ShopifyData[cartItem.PlaceID].Discount = cc.AppliedDiscount
	}
	data.DB.Cart.Save(cart)

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
