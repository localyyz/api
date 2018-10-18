package deals

import (
	"context"
	"net/http"
	"strconv"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/activate", ActivateDeal)

	// TODO remove + backwards compat
	r.With(StatusCtx(data.DealStatusQueued)).Get("/upcoming", ListDeal)
	r.With(StatusCtx(data.DealStatusInactive)).Get("/history", ListDeal)
	r.Route("/active", func(r chi.Router) {
		r.With(StatusCtx(data.DealStatusActive)).Get("/", ListDeal)
		// below is needed needed on the frontend to poll data
		r.Route("/{dealID}", func(r chi.Router) {
			r.Use(DealCtx)
			r.Get("/", GetDeal)
		})
	})

	r.Get("/ongoing", ListOngoingDeal)
	r.Get("/timed", ListTimedDeal)
	r.Get("/comingsoon", ListUpcomingDeal)
	r.Get("/featured", ListFeaturedDeal)
	r.Route("/{dealID}", func(r chi.Router) {
		r.Use(DealCtx)
		r.Get("/", GetDeal)
		r.Route("/products", api.FilterRoutes(ListProducts))
	})

	return r
}

func StatusCtx(status data.DealStatus) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), DealStatusCtxKey, status))
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

/*
	parses the dealID from the request url and fetches the deal to put in context
*/
func DealCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		dealID, err := strconv.ParseInt(chi.URLParam(r, "dealID"), 10, 64)
		if err != nil {
			render.Render(w, r, api.ErrBadID)
			return
		}
		deal, err := data.DB.Deal.FindOne(
			db.Cond{
				"id": dealID,
			},
		)
		if err != nil {
			render.Respond(w, r, err)
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, "deal", deal)
		lg.SetEntryField(ctx, "deal_id", deal.ID)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}
