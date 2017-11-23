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
	*data.CartItem

	Color string `json:"color"`
	Size  string `json:"size"`

	ID     interface{} `json:"id"`
	CartID interface{} `json:"cartId"`

	CreatedAt interface{} `json:"createdAt"`
	UpdatedAt interface{} `json:"updatedAt"`
	DeletedAt interface{} `json:"deletedAt"`
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
	var variant *data.ProductVariant
	err := data.DB.ProductVariant.Find(
		db.And(
			db.Cond{"product_id": payload.ProductID, "limits >=": "1"},
			db.Raw("lower(etc->>'color') = ?", payload.Color),
			db.Raw("lower(etc->>'size') = ?", payload.Size),
		),
	).One(&variant)
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

	if cart.Status == data.CartStatusCheckout {
		// push the change to shopify
		creds, err := data.DB.ShopifyCred.FindByPlaceID(cartItem.PlaceID)
		if err != nil {
			render.Respond(w, r, err)
			return
		}

		cl := shopify.NewClient(nil, creds.AccessToken)
		cl.BaseURL, _ = url.Parse(creds.ApiURL)

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

		var token string
		if sh, ok := cart.Etc.ShopifyData[cartItem.PlaceID]; ok {
			token = sh.Token
		}

		cc, _, _ := cl.Checkout.Update(
			ctx,
			&shopify.CheckoutRequest{
				Checkout: &shopify.Checkout{
					Token:     token,
					LineItems: lineItems,
				},
			},
		)
		cart.Etc.ShopifyData[cartItem.PlaceID] = &data.CartShopifyData{
			TotalTax:   atoi(cc.TotalTax),
			TotalPrice: atoi(cc.TotalPrice),
			PaymentDue: cc.PaymentDue,
			Discount:   cc.AppliedDiscount,
		}
		data.DB.Cart.Save(cart)

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
