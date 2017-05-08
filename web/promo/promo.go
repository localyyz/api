package promo

import (
	"context"
	"net/http"
	"strconv"

	"upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/pressly/chi"
)

const ClaimableDistance = 100.0

func PromoCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		promoID, err := strconv.ParseInt(chi.URLParam(r, "promoID"), 10, 64)
		if err != nil {
			ws.Respond(w, http.StatusBadRequest, api.ErrBadID)
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

func ListHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUser := ctx.Value("session.user").(*data.User)

	claims, err := data.DB.Claim.FindAll(
		db.Cond{
			"user_id": currentUser.ID,
			"status": []data.ClaimStatus{
				data.ClaimStatusCompleted,
				data.ClaimStatusExpired,
			},
		},
	)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	promoIDs := make([]int64, len(claims))
	for i, c := range claims {
		promoIDs[i] = c.PromoID
	}

	var promos []*data.Promo
	err = data.DB.Promo.
		Find(db.Cond{"id": promoIDs}).
		All(&promos)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	res := make([]*presenter.Promo, len(promos))
	for i, p := range promos {
		res[i] = presenter.NewPromo(ctx, p).WithPlace()
	}

	ws.Respond(w, http.StatusOK, res)
}

func ListActive(w http.ResponseWriter, r *http.Request) {
	currentUser := r.Context().Value("session.user").(*data.User)

	claims, err := data.DB.Claim.FindAll(
		db.Cond{
			"user_id": currentUser.ID,
			"status":  data.ClaimStatusActive,
		},
	)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	promoIDs := make([]int64, len(claims))
	claimMap := make(map[int64]*data.Claim, len(claims))
	for i, c := range claims {
		promoIDs[i] = c.PromoID
		claimMap[c.PromoID] = c
	}

	var promos []*data.Promo
	err = data.DB.Promo.
		Find(db.Cond{"id": promoIDs}).
		OrderBy("end_at").
		All(&promos)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	res := make([]*presenter.Promo, len(promos))
	for i, p := range promos {
		res[i] = presenter.NewPromo(r.Context(), p).WithPlace()
		res[i].Claim = claimMap[p.ID]
	}

	ws.Respond(w, http.StatusOK, res)
}
