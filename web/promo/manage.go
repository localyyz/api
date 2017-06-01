package promo

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
	"github.com/pressly/chi/render"

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
		OrderBy("-end_at").All(&promos)
	if err != nil {
		e := errors.Wrap(err, "unable to find access promos")
		ws.Respond(w, http.StatusInternalServerError, e)
		return
	}

	presented := presenter.NewPromoList(ctx, promos)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Render(w, r, nil)
		return
	}
}

func CreatePromo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)
	access := ctx.Value("access").([]*data.UserAccess)

	var promoWrapper struct {
		data.Promo

		ID     interface{} `json:"id,omitempty"`
		UserID interface{} `json:"userId,omitempty"`
		Status interface{} `json:"status,omitempty"`

		CreatedAt interface{} `json:"createdAt,omitempty"`
		UpdatedAt interface{} `json:"updatedAt,omitempty"`
		DeletedAt interface{} `json:"deletedAt,omitempty"`
	}
	if err := ws.Bind(r.Body, &promoWrapper); err != nil {
		ws.Respond(w, http.StatusBadRequest, err)
		return
	}
	newPromo := &promoWrapper.Promo
	newPromo.UserID = user.ID

	// check if placeID is in one of the access
	var allowed bool
	for _, a := range access {
		if newPromo.PlaceID == a.PlaceID {
			allowed = true
			break
		}
	}
	if !allowed {
		ws.Respond(w, http.StatusForbidden, "")
		return
	}

	if err := data.DB.Promo.Save(newPromo); err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	ws.Respond(w, http.StatusCreated, newPromo)
}

func UpdatePromo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	access := ctx.Value("access").([]*data.UserAccess)
	promo := ctx.Value("promo").(*data.Promo)

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

	promoWrapper := struct {
		*data.Promo

		ID      interface{} `json:"id,omitempty"`
		UserID  interface{} `json:"userId,omitempty"`
		PlaceID interface{} `json:"placeId,omitempty"`
		Status  interface{} `json:"status,omitempty"`

		CreatedAt interface{} `json:"createdAt,omitempty"`
		UpdatedAt interface{} `json:"updatedAt,omitempty"`
		DeletedAt interface{} `json:"deletedAt,omitempty"`
	}{Promo: promo}
	if err := ws.Bind(r.Body, &promoWrapper); err != nil {
		ws.Respond(w, http.StatusBadRequest, err)
		return
	}

	if err := data.DB.Promo.Save(promo); err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	ws.Respond(w, http.StatusOK, promo)
}

func DeletePromo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	access := ctx.Value("access").([]*data.UserAccess)
	promo := ctx.Value("promo").(*data.Promo)

	// only allow deleting promotion that's in scheduled state
	if promo.Status != data.PromoStatusScheduled {
		ws.Respond(w, http.StatusBadRequest, errors.New("promotion cannot be active"))
		return
	}

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

	promo.DeletedAt = data.GetTimeUTCPointer()
	promo.Status = data.PromoStatusDeleted
	if err := data.DB.Promo.Save(promo); err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}
	ws.Respond(w, http.StatusNoContent, "")
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

	presented := presenter.NewPromo(ctx, &promo)
	if err := render.Render(w, r, presented); err != nil {
		render.Render(w, r, nil)
		return
	}
}
