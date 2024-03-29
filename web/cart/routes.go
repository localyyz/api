package cart

import (
	"context"
	"net/http"
	"strconv"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"bitbucket.org/moodie-app/moodie-api/web/cart/cartitem"
	"bitbucket.org/moodie-app/moodie-api/web/cart/express"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Mount("/express", express.Routes())
	r.Route("/default", func(r chi.Router) {
		r.Use(DefaultCartCtx)
		r.Mount("/", cartRoutes())
	})
	r.Route("/{cartID}", func(r chi.Router) {
		r.Use(CartCtx)
		r.Mount("/", cartRoutes())
	})
	//r.With(CartStatusScopeCtx).Get("/{cartStatus}", ListCart)

	return r
}

func cartRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", GetCart)
	r.Put("/", UpdateCart)
	r.Delete("/", ClearCart)

	r.Delete("/shipping", DeleteCartShipping)
	r.Delete("/billing", DeleteCartBilling)

	r.Mount("/items", cartitem.Routes())
	r.Get("/shipping", ListShippingRates)

	r.Post("/checkout", CreateCheckouts)
	r.With(CheckoutCtx).Put("/checkout/{checkoutID}", UpdateCheckout)
	r.Post("/pay", CreatePayments)

	return r
}

func CartStatusScopeCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		rawCartStatus := chi.URLParam(r, "cartStatus")
		cartStatus := new(data.CartStatus)
		if err := cartStatus.UnmarshalText([]byte(rawCartStatus)); err != nil {
			render.Respond(w, r, api.ErrInvalidRequest(err))
			return
		}
		//putting a db.Cond in context instead of cartStatus since its the safest
		cartCond := db.Cond{"status": cartStatus}
		ctx = context.WithValue(ctx, "cart.status.scope", cartCond)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func DefaultCartCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := ctx.Value("session.user").(*data.User)

		var cart *data.Cart
		err := data.DB.Cart.Find(
			db.Cond{
				"status <=":  data.CartStatusCheckout,
				"is_express": false,
				"user_id":    user.ID,
			},
		).OrderBy("-id").One(&cart)
		if err != nil {
			if err != db.ErrNoMoreRows {
				render.Respond(w, r, err)
				return
			}
			cart = &data.Cart{
				UserID: user.ID,
				Status: data.CartStatusInProgress,
			}
			data.DB.Cart.Save(cart)
		}

		ctx = context.WithValue(ctx, "cart", cart)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

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
				"id":         cartID,
				"is_express": false,
				"user_id":    user.ID,
			},
		)
		if err != nil {
			render.Respond(w, r, err)
			return
		}
		ctx = context.WithValue(ctx, "cart", cart)
		lg.SetEntryField(ctx, "cart_id", cart.ID)
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
