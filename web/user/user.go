package user

import (
	"context"
	"net/http"
	"strconv"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
	"bitbucket.org/moodie-app/moodie-api/web/api"

	"github.com/pressly/chi"
)

func MeCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		me, ok := ctx.Value("session.user").(*data.User)
		if !ok {
			ws.Respond(w, http.StatusUnauthorized, nil)
			return
		}
		ctx = context.WithValue(ctx, "user", me)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func UserCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
		if err != nil {
			ws.Respond(w, http.StatusBadRequest, api.ErrBadID)
			return
		}

		user, err := data.DB.User.FindByID(userID)
		if err != nil {
			ws.Respond(w, http.StatusInternalServerError, err)
			return
		}
		ctx = context.WithValue(ctx, "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*data.User)
	ws.Respond(w, http.StatusOK, user)
}

func SetDeviceToken(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("session.user").(*data.User)

	var payload struct {
		DeviceToken string `json:"deviceToken,required"`
	}
	if err := ws.Bind(r.Body, &payload); err != nil {
		ws.Respond(w, http.StatusBadRequest, err)
		return
	}

	user.DeviceToken = &payload.DeviceToken
	if err := data.DB.User.Save(user); err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	ws.Respond(w, http.StatusOK, user)
}
