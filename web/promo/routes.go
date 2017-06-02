package promo

import (
	"context"
	"net/http"
	"strconv"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/pressly/chi"
	"github.com/pressly/chi/render"
)

func PromoCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		promoID, err := strconv.ParseInt(chi.URLParam(r, "promoID"), 10, 64)
		if err != nil {
			render.Render(w, r, api.ErrInvalidRequest(err))
			return
		}

		promo, err := data.DB.Promo.FindByID(promoID)
		if err != nil {
			render.Render(w, r, api.WrapErr(err))
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, "promo", promo)
		next.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(handler)
}

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Route("/:promoID", func(r chi.Router) {
		r.Use(PromoCtx)

		r.Route("/claims", func(r chi.Router) {
			r.Use(ClaimCtx)
			r.Get("/", GetClaims)
			r.Put("/complete", CompleteClaim)
			r.Delete("/", RemoveClaim)
		})
	})

	return r
}
