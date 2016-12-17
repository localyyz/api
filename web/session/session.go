package session

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
)

func SessionCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// check Authorization header for jwt
		auth := r.Header.Get("Authorization")
		if auth == "" {
			ws.Respond(w, http.StatusUnauthorized, errors.New("no authorization header"))
			return
		}

		const prefix = "BEARER "
		if !strings.HasPrefix(auth, prefix) {
			ws.Respond(w, http.StatusUnauthorized, errors.New("invalid authorization header"))
			return
		}

		user, err := data.NewSessionUser(auth[len(prefix):])
		if err != nil {
			ws.Respond(w, http.StatusUnauthorized, err)
			return
		}
		ctx = context.WithValue(ctx, "session.user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

// VerifySession can be used to do session verification
func VerifySession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)

	var challenge data.User
	if err := ws.Bind(r.Body, &challenge); err != nil {
		ws.Respond(w, http.StatusBadRequest, err)
		return
	}

	if user.IsAdmin != challenge.IsAdmin {
		ws.Respond(w, http.StatusBadRequest, "")
		return
	}

	ws.Respond(w, http.StatusOK, map[string]bool{"success": true})
}

func Logout(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("session.user").(*data.User)

	// logout the user
	user.LoggedIn = false
	if err := data.DB.User.Save(user); err != nil {
		ws.Respond(w, http.StatusServiceUnavailable, err)
		return
	}

	ws.Respond(w, http.StatusNoContent, "")
}
