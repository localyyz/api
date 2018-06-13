package cart

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/pkg/errors"

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
	Email           string            `json:"email,omitempty"`
}

func (c *cartRequest) Bind(r *http.Request) error {
	return nil
}

func UpdateCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cart := ctx.Value("cart").(*data.Cart)

	if cart.Status > data.CartStatusCheckout {
		err := errors.New("invalid cart status")
		render.Respond(w, r, api.ErrInvalidRequest(err))
		return
	}

	var payload cartRequest
	if err := render.Bind(r, &payload); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	if payload.ShippingAddress != nil {
		cart.ShippingAddress = payload.ShippingAddress
	}
	if payload.BillingAddress != nil {
		cart.BillingAddress = payload.BillingAddress
	}
	if len(payload.Email) != 0 {
		cart.Email = payload.Email
	}

	cart.Status = data.CartStatusInProgress
	if err := data.DB.Cart.Save(cart); err != nil {
		render.Respond(w, r, err)
		return
	}
	render.Render(w, r, presenter.NewCart(ctx, cart))
}

func ClearCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cart := ctx.Value("cart").(*data.Cart)

	if cart.Status > data.CartStatusCheckout {
		err := errors.New("invalid cart status")
		render.Respond(w, r, api.ErrInvalidRequest(err))
		return
	}

	err := data.DB.CartItem.Find(db.Cond{"cart_id": cart.ID}).Delete()
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	cart.Status = data.CartStatusInProgress
	if err := data.DB.Save(cart); err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Status(r, http.StatusNoContent)
	render.Respond(w, r, "")
}

func DeleteCartShipping(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cart := ctx.Value("cart").(*data.Cart)

	if cart.Status > data.CartStatusCheckout {
		err := errors.New("invalid cart status")
		render.Respond(w, r, api.ErrInvalidRequest(err))
		return
	}

	cart.ShippingAddress = nil
	if err := data.DB.Save(cart); err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Respond(w, r, presenter.NewCart(ctx, cart))
}

func DeleteCartBilling(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cart := ctx.Value("cart").(*data.Cart)

	if cart.Status > data.CartStatusCheckout {
		err := errors.New("invalid cart status")
		render.Respond(w, r, api.ErrInvalidRequest(err))
		return
	}

	cart.BillingAddress = nil
	if err := data.DB.Save(cart); err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Respond(w, r, presenter.NewCart(ctx, cart))
}
