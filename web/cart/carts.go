package cart

import (
	"net/http"

	"github.com/go-chi/render"

	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
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
	ctx := r.Context()
	cart := ctx.Value("cart").(*data.Cart)
	render.Render(w, r, presenter.NewCart(ctx, cart))
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
