package session

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	db "upper.io/db.v3"

	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
	"github.com/pkg/errors"
	"github.com/pressly/lg"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/web/api"
)

func SessionCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		token, claims, err := jwtauth.FromContext(ctx)
		if token == nil || err != nil {
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		rawUserID, ok := claims["user_id"].(json.Number)
		if !ok {
			lg.Error("invalid session token, no user_id found")
			render.Render(w, r, api.ErrInvalidSession)
			return
		}

		userID, err := rawUserID.Int64()
		if err != nil {
			lg.Errorf("invalid session token: %+v", err)
			render.Render(w, r, api.ErrInvalidSession)
			return
		}

		// find a logged in user with the given id
		user, err := data.DB.User.FindOne(
			db.Cond{
				"id":        userID,
				"logged_in": true,
			},
		)
		if err != nil {
			lg.Errorf("invalid session user: %+v", err)
			render.Render(w, r, api.ErrInvalidSession)
			return
		}

		ctx = context.WithValue(ctx, "session.user", user)
		lg.SetEntryField(ctx, "user_id", user.ID)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

// Device context uses device id to create an unique token that represents a new
// user account
func DeviceCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if _, ok := ctx.Value("session.user").(*data.User); ok {
			// Use already exists. Nothing to do
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		deviceId := r.Header.Get("X-DEVICE-ID")
		if deviceId == "" {
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// find a user with username or device token of deviceID
		var user *data.User
		res := data.DB.User.Find(
			db.Or(
				db.Cond{"username": deviceId},
				db.Cond{"device_token": deviceId},
			),
		)
		err := res.One(&user)
		if err != nil {
			// any error other than a row not found
			if err != db.ErrNoMoreRows {
				render.Respond(w, r, err)
				return
			}

			// did not find user by either username or device token
			// so create new shadow user and save it
			user = &data.User{
				Username: deviceId,
				Email:    deviceId,
				Network:  "shadow",
			}
		}

		// update last time user was logged in
		user.LastLogInAt = data.GetTimeUTCPointer()
		user.LoggedIn = true

		// save the new user
		data.DB.User.Save(user)

		lg.SetEntryField(ctx, "user_id", user.ID)
		ctx = context.WithValue(ctx, "session.user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func UserRefresh(next http.Handler) http.Handler {
	// UserRefresh periodically refreshes user data from their
	// social network
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if user, ok := ctx.Value("session.user").(*data.User); ok {
			// for now, just facebook
			if user.Network == "facebook" {
				if user.UpdatedAt == nil || time.Since(*user.UpdatedAt) > 20*24*time.Hour {
					if err := connect.FacebookLogin.GetUser(user); err != nil {
						lg.Warn(errors.Wrap(err, "unable to refresh user"))
					} else {
						data.DB.User.Save(user)
					}
				}
			}
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(handler)

}

func Logout(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("session.user").(*data.User)

	// logout the user
	user.LoggedIn = false
	if err := data.DB.User.Save(user); err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Status(r, http.StatusNoContent)
	render.Respond(w, r, "")
}
