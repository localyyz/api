package promo

import (
	"context"
	"net/http"

	"github.com/pkg/errors"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
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
			if err == db.ErrNoMoreRows {
				ws.Respond(w, http.StatusNotFound, "")
				return
			}
			ws.Respond(w, http.StatusInternalServerError, errors.Wrap(err, "failed to query claim"))
			return
		}

		ctx = context.WithValue(ctx, "claim", claim)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func GetClaims(w http.ResponseWriter, r *http.Request) {
	claim := r.Context().Value("claim").(*data.Claim)
	ws.Respond(w, http.StatusOK, claim)
}

func CompleteClaim(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claim := ctx.Value("claim").(*data.Claim)

	claim.Status = data.ClaimStatusCompleted
	if err := data.DB.Claim.Save(claim); err != nil {
		ws.Respond(w, http.StatusInternalServerError, errors.Wrap(err, "complete"))
		return
	}
	ws.Respond(w, http.StatusOK, claim)
}

func RemoveClaim(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claim := ctx.Value("claim").(*data.Claim)

	claim.Status = data.ClaimStatusRemoved
	if err := data.DB.Claim.Save(claim); err != nil {
		ws.Respond(w, http.StatusInternalServerError, errors.Wrap(err, "remove"))
		return
	}
	ws.Respond(w, http.StatusNoContent, "")
}
