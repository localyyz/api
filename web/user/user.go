package user

import (
	"context"
	"net/http"

	"github.com/pressly/chi/render"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
	"bitbucket.org/moodie-app/moodie-api/web/api"
)

func MeCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		me, ok := ctx.Value("session.user").(*data.User)
		if !ok {
			render.Render(w, r, api.ErrInvalidSession)
			return
		}
		ctx = context.WithValue(ctx, "user", me)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func AcceptNDA(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)

	user.Etc.HasAgreedNDA = true
	if err := data.DB.User.Save(user); err != nil {
		render.Render(w, r, api.WrapErr(err))
		return
	}

	render.Render(w, r, presenter.NewUser(ctx, user))
}

func SetDeviceToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)

	var payload struct {
		DeviceToken string `json:"deviceToken,required"`
	}
	if err := ws.Bind(r.Body, &payload); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	user.DeviceToken = &payload.DeviceToken
	if err := data.DB.User.Save(user); err != nil {
		render.Render(w, r, api.WrapErr(err))
		return
	}

	render.Render(w, r, presenter.NewUser(ctx, user))
}
