package session

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	db "upper.io/db.v3"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"github.com/pressly/chi/render"
	"github.com/pressly/lg"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/token"
	"bitbucket.org/moodie-app/moodie-api/web/api"
)

func SessionCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tok, _ := ctx.Value("jwt").(*jwt.Token)
		if tok == nil {
			render.Render(w, r, api.ErrInvalidSession)
			return
		}

		token, err := token.Decode(tok.Raw)
		if err != nil {
			lg.Errorf("invalid session token: %+v", err)
			render.Render(w, r, api.ErrInvalidSession)
			return
		}

		rawUserID, ok := token.Claims["user_id"].(json.Number)
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
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func UserRefresh(next http.Handler) http.Handler {
	// UserRefresh periodically refreshes user data from their
	// social network
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := ctx.Value("session.user").(*data.User)

		// for now, just facebook
		if user.Network == "facebook" {
			if user.UpdatedAt == nil || time.Since(*user.UpdatedAt) > 20*24*time.Hour {
				if err := connect.FB.GetUser(user); err != nil {
					lg.Warn(errors.Wrap(err, "unable to refresh user"))
					return
				}
				data.DB.User.Save(user)
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
