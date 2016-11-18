package promo

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pressly/chi"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	db "upper.io/db.v2"
)

func PromoActionCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		action := chi.URLParam(r, "action")

		// handle appropriate claim action
		var status data.ClaimStatus
		switch action {
		case "claim":
			status = data.ClaimStatusActive
		case "save":
			status = data.ClaimStatusSaved
		case "peek":
			status = data.ClaimStatusPeeked
		default:
			ws.Respond(w, http.StatusNotFound, "")
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, "status", status)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func DoPromo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	promo := ctx.Value("promo").(*data.Promo)
	status := ctx.Value("status").(data.ClaimStatus)
	currentUser := ctx.Value("session.user").(*data.User)

	// if we're claiming the promotion, check the distance
	if status == data.ClaimStatusActive {
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
			ws.Respond(w, http.StatusBadRequest, api.ErrClaimDistance)
			return
		}
	}

	claim := &data.Claim{
		PromoID: promo.ID,
		PlaceID: promo.PlaceID,
		UserID:  currentUser.ID,
		Status:  status,
	}
	if err := data.DB.Claim.Save(claim); err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	ws.Respond(w, http.StatusCreated, claim)
}
