package promo

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
	db "upper.io/db.v2"
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
		}).One(&claim)
		if err != nil && err != db.ErrNoMoreRows {
			ws.Respond(w, http.StatusInternalServerError, errors.Wrap(err, "failed to query claim"))
			return
		}

		ctx = context.WithValue(ctx, "claim", claim)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func ClaimPromo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claim := ctx.Value("claim").(*data.Claim)
	promo := ctx.Value("promo").(*data.Promo)
	currentUser := ctx.Value("session.user").(*data.User)

	// already claimed
	if claim != nil {
		if claim.Status == data.ClaimStatusActive {
			ws.Respond(w, http.StatusOK, claim)
			return
		}

		// previously claimed
		if claim.Status == data.ClaimStatusExpired ||
			claim.Status == data.ClaimStatusCompleted {
			ws.Respond(w, http.StatusConflict, errors.New("promotion already claimed"))
			return
		}
	}

	// update the status
	claim = &data.Claim{
		PromoID: promo.ID,
		PlaceID: promo.PlaceID,
		UserID:  currentUser.ID,
		Status:  data.ClaimStatusActive,
	}

	// calculate the user's distance from the "place"
	var place *data.Place
	err := data.DB.Place.Find(
		db.Cond{"id": promo.PlaceID},
	).Select(
		db.Raw(fmt.Sprintf("ST_Distance(geo, st_geographyfromtext('%v'::text)) distance", currentUser.Geo)),
	).One(&place)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	if place.Distance > ClaimableDistance {
		//ws.Respond(w, http.StatusBadRequest, api.ErrClaimDistance)
		//return
		// NOTE: if out of distance, auto save
		claim.Status = data.ClaimStatusSaved
	}

	if err := data.DB.Claim.Save(claim); err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	ws.Respond(w, http.StatusCreated, claim)
}

func SavePromo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claim := ctx.Value("claim").(*data.Claim)
	promo := ctx.Value("promo").(*data.Promo)
	currentUser := ctx.Value("session.user").(*data.User)

	if claim != nil {
		// can't save already existing claim
		ws.Respond(w, http.StatusBadRequest, errors.New("already saved/claimed"))
		return
	}

	// update the status
	claim = &data.Claim{
		PromoID: promo.ID,
		PlaceID: promo.PlaceID,
		UserID:  currentUser.ID,
		Status:  data.ClaimStatusSaved,
	}

	if err := data.DB.Claim.Save(claim); err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	ws.Respond(w, http.StatusCreated, claim)
}

func UnSavePromo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claim := ctx.Value("claim").(*data.Claim)

	if claim.Status != data.ClaimStatusSaved {
		ws.Respond(w, http.StatusBadRequest, errors.New("unable to remove claimed promotion."))
		return
	}

	if err := data.DB.Claim.Delete(claim); err != nil {
		ws.Respond(w, http.StatusInternalServerError, errors.Wrap(err, "remove promotion db error"))
		return
	}

	ws.Respond(w, http.StatusNoContent, "")
}
