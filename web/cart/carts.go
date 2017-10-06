package cart

import (
	"net/http"

	db "upper.io/db.v3"

	"github.com/pressly/chi/render"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
)

func ListCarts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)
	scope, ok := ctx.Value("scope").(db.Cond)
	if !ok {
		scope = db.Cond{}
	}
	scope["user_id"] = user.ID

	var carts []*data.Cart
	err := data.DB.Cart.Find(scope).OrderBy("status").All(&carts)
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
