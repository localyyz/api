package cart

import (
	"context"
	"net/http"
	"strconv"

	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/goware/lg"
	"github.com/pressly/chi"
	"github.com/pressly/chi/render"
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

func CartItemCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		cartID, err := strconv.ParseInt(chi.URLParam(r, "cartID"), 10, 64)
		if err != nil {
			render.Render(w, r, api.ErrBadID)
			return
		}
		cartItemID, err := strconv.ParseInt(chi.URLParam(r, "cartItemID"), 10, 64)
		if err != nil {
			render.Render(w, r, api.ErrBadID)
			return
		}
		ctx := r.Context()
		user := ctx.Value("session.user").(*data.User)

		cartItem, err := data.DB.CartItem.FindOne(
			db.Cond{
				"id":      cartItemID,
				"cart_id": cartID,
				"user_id": user.ID,
			},
		)
		if err != nil {
			render.Respond(w, r, err)
			return
		}
		ctx = context.WithValue(ctx, "cart_item", cartItem)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
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
		Quantity:  uint32(payload.Quantity),
	}

	if err := data.DB.CartItem.Save(newItem); err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Render(w, r, presenter.NewCartItem(ctx, newItem))
}

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

	if err := data.DB.CartItem.Delete(cartItem); err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Status(r, http.StatusNoContent)
	render.Respond(w, r, "")
}
