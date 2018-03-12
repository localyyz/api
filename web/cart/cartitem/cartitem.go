package cartitem

import (
	"net/http"
	"strconv"

	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
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

func CreateCartItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cart := ctx.Value("cart").(*data.Cart)

	var payload cartItemRequest
	if err := render.Bind(r, &payload); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	// Reset cart status to inProgress
	if cart.Status != data.CartStatusInProgress {
		cart.Status = data.CartStatusInProgress
		if err := data.DB.Cart.Save(cart); err != nil {
			render.Render(w, r, api.ErrInvalidRequest(err))
			return
		}
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
		if err == db.ErrNoMoreRows {
			render.Render(w, r, api.ErrOutOfStockAdd(err))
			return
		}
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

func RemoveCartItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cart := ctx.Value("cart").(*data.Cart)
	cartItem := ctx.Value("cart_item").(*data.CartItem)

	// check if this is the last item in cart -> shopifyData [ merchant ]
	numItems, err := data.DB.CartItem.Find(
		db.Cond{
			"cart_id":  cart.ID,
			"place_id": cartItem.PlaceID,
			"id <>":    cartItem.ID,
		},
	).Count()
	if err != nil {
		render.Respond(w, r, api.ErrInvalidRequest(err))
		return
	}

	if numItems == 0 {
		// if this is the last item from this merchant. remove the checkout from shopify data
		delete(cart.Etc.ShopifyData, cartItem.PlaceID)
		delete(cart.Etc.ShippingMethods, cartItem.PlaceID)
	}

	// Remove the cart item
	if err := data.DB.CartItem.Delete(cartItem); err != nil {
		render.Respond(w, r, api.ErrInvalidRequest(err))
		return
	}

	// Reset cart status to inProgress
	if cart.Status != data.CartStatusInProgress {
		cart.Status = data.CartStatusInProgress
		if err := data.DB.Cart.Save(cart); err != nil {
			render.Render(w, r, api.ErrInvalidRequest(err))
			return
		}
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
