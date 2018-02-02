package express

import (
	"context"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	db "upper.io/db.v3"
)

func ExpressCartCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := ctx.Value("session.user").(*data.User)

		var cart *data.Cart
		err := data.DB.Cart.Find(
			db.Cond{
				"status <=":  data.CartStatusCheckout,
				"is_express": true,
				"user_id":    user.ID,
			},
		).OrderBy("-id").One(&cart)
		if err != nil {
			if err != db.ErrNoMoreRows {
				render.Respond(w, r, err)
				return
			}
			cart = &data.Cart{
				UserID:    user.ID,
				IsExpress: true,
				Status:    data.CartStatusInProgress,
			}
			data.DB.Cart.Save(cart)
		}

		ctx = context.WithValue(ctx, "cart", cart)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(ExpressCartCtx)

	r.Post("/items", CreateCartItem)
	r.Post("/pay", CreatePayment)

	return r
}
