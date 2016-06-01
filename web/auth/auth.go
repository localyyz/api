package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/pressly/chi"

	"bitbucket.org/pxue/api/data"
	"bitbucket.org/pxue/api/lib/connect"
	"bitbucket.org/pxue/api/lib/ws"

	"golang.org/x/net/context"
)

func SessionCtx(next chi.Handler) chi.Handler {
	return chi.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {

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

		next.ServeHTTPC(ctx, w, r)
	})
}

// FacebookLogin handles both first-time login (signup) and repeated-logins from a social network
// User is already authenticated by the frontend with network of their choice
//  Backend stores the token and async grab the user data
func FacebookLogin(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Token string `json:"token,required"`
	}
	if err := ws.Bind(r.Body, &payload); err != nil {
		ws.Respond(w, http.StatusBadRequest, err)
		return
	}

	// inspect the token for userId and expiry
	user, err := connect.FB.Login(payload.Token)
	if err != nil {
		if err == connect.ErrTokenExpired {
			ws.Respond(w, http.StatusUnauthorized, errors.New("token expired"))
			return
		}
		ws.Respond(w, http.StatusServiceUnavailable, err)
		return
	}

	authUser, err := data.NewAuthUser(user)
	if err != nil {
		ws.Respond(w, http.StatusUnauthorized, err)
		return
	}

	ws.Respond(w, http.StatusOK, authUser)
}

func Logout(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	user := ctx.Value("session.user").(*data.Account)

	// logout the user
	user.LoggedIn = false
	if err := data.DB.Account.Save(user); err != nil {
		ws.Respond(w, http.StatusServiceUnavailable, err)
		return
	}

	ws.Respond(w, http.StatusNoContent, "")
}
