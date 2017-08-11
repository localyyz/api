package cart

import (
	"net/http"

	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/goware/lg"
	"github.com/pressly/chi/render"
)

type cartItemRequest struct {
	ProductID int64  `json:"productId"`
	Color     string `json:"color"`
	Size      string `json:"size"`
	VariantID int64  `json:"variantId,omitempty"`
	Quantity  int64  `json:"quantity"`
}

func (*cartItemRequest) Bind(r *http.Request) error {
	return nil
}

func CartItemCtx(next http.Handler) http.Handler {
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
			db.Cond{"product_id": payload.ProductID},
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
}

func RemoveCartItem(w http.ResponseWriter, r *http.Request) {
}
