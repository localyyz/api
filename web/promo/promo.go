package promo

import (
	"context"
	"net/http"
	"strconv"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
	"bitbucket.org/moodie-app/moodie-api/web/utils"
	"github.com/pressly/chi"
)

func PromoCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		promoID, err := strconv.ParseInt(chi.URLParam(r, "promoID"), 10, 64)
		if err != nil {
			ws.Respond(w, http.StatusBadRequest, utils.ErrBadID)
			return
		}

		// TODO: check if promo is within distance
		// TODO: get by hash

		promo, err := data.DB.Promo.FindByID(promoID)
		if err != nil {
			ws.Respond(w, http.StatusInternalServerError, err)
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, "promo", promo)
		next.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(handler)
}

func GetPromo(w http.ResponseWriter, r *http.Request) {
	promo := r.Context().Value("promo").(*data.Promo)
	ws.Respond(w, http.StatusOK, promo)
}

func ClaimPromo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	promo := ctx.Value("promo").(*data.Promo)
	currentUser := ctx.Value("session.user").(*data.User)

	claim := &data.Claim{
		PromoID: promo.ID,
		PlaceID: promo.PlaceID,
		UserID:  currentUser.ID,
		Status:  data.ClaimStatusActive,
	}
	if err := data.DB.Claim.Save(claim); err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	ws.Respond(w, http.StatusCreated, claim)
}
