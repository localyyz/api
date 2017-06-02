package promo

import (
	"context"
	"net/http"

	"github.com/pressly/chi/render"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"upper.io/db.v3"
)

func ClaimCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		currentUser := ctx.Value("session.user").(*data.User)
		promo := ctx.Value("promo").(*data.Promo)

		var claim *data.Claim
		err := data.DB.Claim.Find(db.Cond{
			"promo_id": promo.ID,
			"user_id":  currentUser.ID,
			"status":   data.ClaimStatusActive,
		}).One(&claim)
		if err != nil {
			render.Render(w, r, api.WrapErr(err))
			return
		}

		ctx = context.WithValue(ctx, "claim", claim)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func GetClaims(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claim := ctx.Value("claim").(*data.Claim)
	render.Render(w, r, presenter.NewClaim(ctx, claim))
}

func CompleteClaim(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claim := ctx.Value("claim").(*data.Claim)

	claim.Status = data.ClaimStatusCompleted
	if err := data.DB.Claim.Save(claim); err != nil {
		render.Render(w, r, api.WrapErr(err))
		return
	}
	render.Render(w, r, presenter.NewClaim(ctx, claim))
}

func RemoveClaim(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claim := ctx.Value("claim").(*data.Claim)
	claim.Status = data.ClaimStatusRemoved
	if err := data.DB.Claim.Save(claim); err != nil {
		render.Render(w, r, api.WrapErr(err))
		return
	}
	render.Render(w, r, api.NoContentResp)
}
