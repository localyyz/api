package cart

import (
	"net/http"

	"github.com/go-chi/render"

	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
)

func GetCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cart := ctx.Value("cart").(*data.Cart)
	render.Render(w, r, presenter.NewCart(ctx, cart))
}

type cartRequest struct {
	ShippingAddress *data.CartAddress `json:"shippingAddress,omitempty"`
	BillingAddress  *data.CartAddress `json:"billingAddress,omitempty"`
	//Shipping        *data.CartShippingMethod `json:"shipping,omitempty"`
	DiscountCode string `json:"discountCode,omitempty"`
}

func (c *cartRequest) Bind(r *http.Request) error {
	return nil
}

func UpdateCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cart := ctx.Value("cart").(*data.Cart)

	var payload cartRequest
	if err := render.Bind(r, &payload); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	if a := payload.ShippingAddress; a != nil {
		cart.Etc.ShippingAddress = a
	}
	if b := payload.BillingAddress; b != nil {
		cart.Etc.BillingAddress = b
	}

	if err := data.DB.Cart.Save(cart); err != nil {
		render.Respond(w, r, err)
		return
	}
	render.Render(w, r, presenter.NewCart(ctx, cart))
}

func ClearCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cart := ctx.Value("cart").(*data.Cart)

	err := data.DB.CartItem.Find(db.Cond{"cart_id": cart.ID}).Delete()
	if err != nil {
		render.Respond(w, r, err)
		return
	}
	render.Status(r, http.StatusNoContent)
	render.Respond(w, r, "")
}
