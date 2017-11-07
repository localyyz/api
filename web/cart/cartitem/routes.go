package cartitem

import (
	"context"
	"net/http"
	"strconv"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	db "upper.io/db.v3"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", CreateCartItem)
	r.Route("/{cartItemID}", func(r chi.Router) {
		r.Use(CartItemCtx)
		r.Get("/", GetCartItem)
		r.Put("/", UpdateCartItem)
		r.Delete("/", RemoveCartItem)
	})

	return r
}

func CartItemCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		cart := ctx.Value("cart").(*data.Cart)
		cartItemID, err := strconv.ParseInt(chi.URLParam(r, "cartItemID"), 10, 64)
		if err != nil {
			render.Render(w, r, api.ErrBadID)
			return
		}

		// by this point, cart ctx should have verified
		// the user ownership
		cartItem, err := data.DB.CartItem.FindOne(
			db.Cond{
				"id":      cartItemID,
				"cart_id": cart.ID,
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
