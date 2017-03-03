package promo

import (
	"context"
	"net/http"

	"github.com/pkg/errors"

	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
)

func PromoManageCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("session.user").(*data.User)

		access, err := data.DB.UserAccess.EditorAccess(user.ID)
		if err != nil {
			e := errors.Wrap(err, "unable to find user access")
			ws.Respond(w, http.StatusInternalServerError, e)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, "access", access)
		next.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(handler)
}

func ListManagable(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	access := ctx.Value("access").([]*data.UserAccess)

	// find the places the user can manage
	var placeIDs []int64
	for _, a := range access {
		placeIDs = append(placeIDs, a.PlaceID)
	}

	var promos []*data.Promo
	err := data.DB.Promo.Find(db.Cond{"place_id": placeIDs}).
		OrderBy("end_at").All(&promos)
	if err != nil {
		e := errors.Wrap(err, "unable to find access promos")
		ws.Respond(w, http.StatusInternalServerError, e)
		return
	}

	res := make([]*presenter.Promo, len(promos))
	for i, p := range promos {
		res[i] = presenter.NewPromo(r.Context(), p).WithPlace()
	}

	ws.Respond(w, http.StatusOK, res)
}

func CreatePromo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)
	access := ctx.Value("access").([]*data.UserAccess)

	var promo data.Promo
	if err := ws.Bind(r.Body, &promo); err != nil {
		ws.Respond(w, http.StatusBadRequest, err)
		return
	}
	promo.UserID = user.ID

	// check if placeID is in one of the access
	var allowed bool
	for _, a := range access {
		if promo.PlaceID == a.PlaceID {
			allowed = true
			break
		}
	}
	if !allowed {
		ws.Respond(w, http.StatusForbidden, "")
		return
	}

	if err := data.DB.Promo.Save(&promo); err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	ws.Respond(w, http.StatusCreated, promo)
}

func PreviewPromo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)
	access := ctx.Value("access").([]*data.UserAccess)

	var promo data.Promo
	if err := ws.Bind(r.Body, &promo); err != nil {
		ws.Respond(w, http.StatusBadRequest, err)
		return
	}
	promo.UserID = user.ID

	// check if placeID is in one of the access
	var allowed bool
	for _, a := range access {
		if promo.PlaceID == a.PlaceID {
			allowed = true
			break
		}
	}
	if !allowed {
		ws.Respond(w, http.StatusForbidden, "")
		return
	}

	resp := presenter.NewPromo(ctx, &promo)
	ws.Respond(w, http.StatusCreated, resp.WithPlace())
}
