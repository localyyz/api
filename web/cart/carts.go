package cart

import (
	"context"
	"net/http"
	"strconv"

	db "upper.io/db.v3"

	"github.com/pressly/chi"
	"github.com/pressly/chi/render"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
)

func CartCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		cartID, err := strconv.ParseInt(chi.URLParam(r, "cartID"), 10, 64)
		if err != nil {
			render.Render(w, r, api.ErrBadID)
			return
		}
		ctx := r.Context()
		user := ctx.Value("session.user").(*data.User)

		cart, err := data.DB.Cart.FindOne(
			db.Cond{
				"id":      cartID,
				"user_id": user.ID,
			},
		)
		if err != nil {
			render.Respond(w, r, err)
			return
		}
		ctx = context.WithValue(ctx, "cart", cart)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func CartScopeCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// get any existing scope
		scope, ok := ctx.Value("scope").(db.Cond)
		if !ok {
			scope = db.Cond{}
		}

		// no scope, return all
		if cartScope := chi.URLParam(r, "scope"); len(cartScope) != 0 {
			var cartStatus data.CartStatus
			if err := cartStatus.UnmarshalText([]byte(cartScope)); err != nil {
				render.Respond(w, r, err)
				return
			}
			scope["carts.status"] = cartStatus
		}

		ctx = context.WithValue(ctx, "scope", scope)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ListCarts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)
	scope, ok := ctx.Value("scope").(db.Cond)

	if !ok {
		scope = db.Cond{"user_id": user.ID}
	}

	carts, err := data.DB.Cart.FindAll(scope)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	presented := presenter.NewUserCartList(ctx, carts)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}
}

func GetCart(w http.ResponseWriter, r *http.Request) {
	cart := r.Context().Value("cart").(*data.Cart)
	render.Render(w, r, presenter.NewCart(r.Context(), cart))
}

func CreateCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)

	newCart := &data.Cart{
		UserID: user.ID,
		Status: data.CartStatusInProgress,
	}

	if err := data.DB.Cart.Save(newCart); err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Render(w, r, presenter.NewCart(ctx, newCart))
}

type cartUpdateRequest struct {
	*data.Cart
	ShippingAddress *data.CartAddress `json:"shippingAddress"`
	BillingAddress  *data.CartAddress `json:"billingAddress"`

	ID     interface{} `json:"id,omitempty"`
	UserID interface{} `json:"userId,omitempty"`
	Status interface{} `json:"status,omitempty"`

	CreatedAt interface{} `json:"createdAt,omitempty"`
	UpdatedAt interface{} `json:"updatedAt,omitempty"`
	DeletedAt interface{} `json:"deletedAt,omitempty"`
}

func (c *cartUpdateRequest) Bind(r *http.Request) error {
	return nil
}

func UpdateCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cart := ctx.Value("cart").(*data.Cart)

	var payload cartUpdateRequest
	if err := render.Bind(r, &payload); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	if addy := payload.ShippingAddress; addy != nil {
		cart.Etc.ShippingAddress = addy
	}

	if err := data.DB.Cart.Save(cart); err != nil {
		render.Respond(w, r, err)
		return
	}
	render.Respond(w, r, presenter.NewCart(ctx, cart))
}

func DeleteCart(w http.ResponseWriter, r *http.Request) {
}
