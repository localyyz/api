package user

import (
	"context"
	"net/http"

	"github.com/pressly/chi/render"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
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

type deviceTokenRequst struct {
	DeviceToken string `json:"deviceToken,required"`
}

func (*deviceTokenRequst) Bind(r *http.Request) error {
	return nil
}

func SetDeviceToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)

	payload := &deviceTokenRequst{}
	if err := render.Bind(r, payload); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	user.DeviceToken = &payload.DeviceToken
	if err := data.DB.User.Save(user); err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Render(w, r, presenter.NewUser(ctx, user))
}
