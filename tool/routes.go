package tool

import (
	"context"
	"net/http"
	"net/url"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/scheduler"
	"bitbucket.org/moodie-app/moodie-api/web/place"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	db "upper.io/db.v3"
)

type Handler struct {
	DB    *data.Database
	Debug bool
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/testdata", InsertTestPurchasableProduct)
	r.Get("/products/score", syncProductImageScores)

	r.Get("/products/update", UpdateCategories)
	r.Get("/products/count", GetMerchantProductCount)
	r.Get("/places/active", ListActive)
	r.Get("/places/permissions", ListPermissions)
	r.Get("/places/social", GetSocialMedia)
	r.Get("/places/pricerules", ListPriceRules)
	r.Route("/places/{placeID}", func(r chi.Router) {
		r.Use(place.PlaceCtx)
		r.Use(func(next http.Handler) http.Handler {
			handler := func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()
				place := ctx.Value("place").(*data.Place)

				cred, err := data.DB.ShopifyCred.FindOne(db.Cond{"place_id": place.ID})
				if err != nil {
					render.Respond(w, r, err)
					return
				}
				client := shopify.NewClient(nil, cred.AccessToken)
				client.BaseURL, _ = url.Parse(cred.ApiURL)

				ctx = context.WithValue(ctx, "shopify.client", client)
				next.ServeHTTP(w, r.WithContext(ctx))
			}
			return http.HandlerFunc(handler)
		})

		r.Get("/discount", ListPriceRule)
		r.Get("/products/count", GetProductCount)

		r.Get("/products", GetProduct)
		r.Post("/products/sync/{externalID}", SyncProduct)
		r.Put("/products/sync", SyncProducts)
		r.Post("/products/sync", SyncProducts)
		r.Post("/products/validate", ValidateSyncProducts)
		r.Delete("/products/sync", CleanupProduct)

		r.Put("/variants/sync", SyncVariants)
	})

	r.Post("/syncer/deal", func(w http.ResponseWriter, r *http.Request) {
		z := scheduler.New(h.DB)
		z.SyncDeals()
	})

	return r
}
