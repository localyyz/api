package express

import (
	"context"
	"net/http"
	"net/url"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/pkg/errors"
	"github.com/pressly/lg"
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

func ExpressShopifyClientCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		cart := ctx.Value("cart").(*data.Cart)

		if len(cart.Etc.ShopifyData) == 0 {
			render.Respond(w, r, api.ErrInvalidRequest(errors.New("cart is empty")))
			return
		}

		var (
			placeID  int64
			checkout *data.CartShopifyData
		)
		for pID, c := range cart.Etc.ShopifyData {
			placeID = pID
			checkout = c
			checkout.PlaceID = pID
			break
		}

		cred, err := data.DB.ShopifyCred.FindByPlaceID(placeID)
		if err != nil {
			lg.Warnf("unable to find cred for place %d", placeID)
			render.Respond(w, r, err)
			return
		}
		client := shopify.NewClient(nil, cred.AccessToken)
		client.BaseURL, _ = url.Parse(cred.ApiURL)
		client.Debug = true

		ctx = context.WithValue(ctx, "shopify.client", client)
		ctx = context.WithValue(ctx, "shopify.checkout", checkout)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(ExpressCartCtx)

	r.Get("/", GetCart)
	r.Delete("/", DeleteCart)
	r.Post("/items", CreateCartItem)

	r.Group(func(r chi.Router) {
		r.Use(ExpressShopifyClientCtx)

		r.Get("/shipping/estimate", GetShippingRates)
		r.Put("/shipping/address", UpdateShippingAddress)
		r.Put("/shipping/method", UpdateShippingMethod)
		r.Post("/pay", CreatePayment)
	})

	return r
}
