package user

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/render"
	"github.com/pressly/lg"

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

// Use this endpoint to cache bust
// TODO: is the proper way reissue the jwt?
// ie bust the jwt and reissue...?
func Ping(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	me := ctx.Value("session.user").(*data.User)

	// last updated at timestamp from cache
	lastCached, err := strconv.ParseInt(r.URL.Query().Get("lu"), 10, 64)
	if err != nil {
		render.Respond(w, r, api.ErrInvalidRequest(err))
		return
	}
	if me.UpdatedAt != nil && me.UpdatedAt.Unix()-lastCached > 0 {
		// tell frontend to bust the cache
		lg.Warnf("busting user(%d) cache. diff: %d", me.ID, me.UpdatedAt.Unix()-lastCached)
		render.Status(r, http.StatusResetContent)
		render.Respond(w, r, ".")
		return
	}

	// no change, return ok
	render.Status(r, http.StatusOK)
	render.Respond(w, r, ".")
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)
	render.Render(w, r, presenter.NewUser(ctx, user))
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
