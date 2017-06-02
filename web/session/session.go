package session

import (
	"context"
	"encoding/json"
	"net/http"

	db "upper.io/db.v3"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/goware/lg"
	"github.com/pressly/chi/render"

	"bitbucket.org/moodie-app/moodie-api/data"
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
		}

		rawUserID, ok := token.Claims["user_id"].(json.Number)
		if !ok {
			lg.Error("invalid session token, no user_id found")
			render.Render(w, r, api.ErrInvalidSession)
		}

		userID, err := rawUserID.Int64()
		if err != nil {
			lg.Errorf("invalid session token: %+v", err)
			render.Render(w, r, api.ErrInvalidSession)
		}

		// find a logged in user with the given id
		user, err := data.DB.User.FindOne(
			db.Cond{
				"id":        userID,
				"logged_in": true,
			},
		)
		if err != nil {
			lg.Error("invalid session user: %+v", err)
			render.Render(w, r, api.ErrInvalidSession)
			return
		}

		ctx = context.WithValue(ctx, "session.user", user)
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
