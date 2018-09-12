package user

import (
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"github.com/go-chi/render"
	db "upper.io/db.v3"
)

func ListOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)
	carts, err := data.DB.Cart.FindAll(db.Cond{
		"status": []data.CartStatus{
			data.CartStatusComplete,
			data.CartStatusPaymentSuccess,
		},
		"user_id": user.ID,
	})
	if err != nil {
		render.Respond(w, r, err)
		return
	}
	render.RenderList(w, r, presenter.NewUserCartList(ctx, carts))
}
