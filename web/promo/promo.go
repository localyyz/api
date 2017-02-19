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

func PreviewPromo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)

	// NOTE: protected route
	if !user.IsAdmin {
		ws.Respond(w, http.StatusNotFound, "")
		return
	}

	var promo data.Promo
	if err := ws.Bind(r.Body, &promo); err != nil {
		ws.Respond(w, http.StatusBadRequest, err)
		return
	}
	promo.UserID = user.ID

	resp := &presenter.Promo{Promo: &promo}
	ws.Respond(w, http.StatusCreated, resp.WithPlace())
}

func CreatePromo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)

	// NOTE: protected route
	if !user.IsAdmin {
		ws.Respond(w, http.StatusNotFound, "")
		return
	}

	var promo data.Promo
	if err := ws.Bind(r.Body, &promo); err != nil {
		ws.Respond(w, http.StatusBadRequest, err)
		return
	}
	promo.UserID = user.ID

	if err := data.DB.Promo.Save(&promo); err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	ws.Respond(w, http.StatusCreated, promo)
}

func GetClaims(w http.ResponseWriter, r *http.Request) {
	promo := r.Context().Value("promo").(*data.Promo)

	ws.Respond(w, http.StatusOK, promo)
}

func ListHistory(w http.ResponseWriter, r *http.Request) {
	currentUser := r.Context().Value("session.user").(*data.User)

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
		res[i] = &presenter.Promo{Promo: p}
		res[i].WithPlace()
	}

	ws.Respond(w, http.StatusOK, res)
}

func ListActive(w http.ResponseWriter, r *http.Request) {
	currentUser := r.Context().Value("session.user").(*data.User)

	claims, err := data.DB.Claim.FindAll(
		db.Cond{
			"user_id": currentUser.ID,
			"status": []data.ClaimStatus{
				data.ClaimStatusActive,
				data.ClaimStatusSaved,
				data.ClaimStatusPeeked,
			},
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
