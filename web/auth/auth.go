package auth

import (
	"errors"
	"net/http"

	"github.com/pressly/chi"

	"bitbucket.org/pxue/api/lib/connect"
	"bitbucket.org/pxue/api/lib/ws"

	"golang.org/x/net/context"
)

func SessionCtx(next chi.Handler) chi.Handler {
	return chi.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
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

	ws.Respond(w, http.StatusOK, user)
}

func Logout(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ws.Respond(w, http.StatusOK, "")
}
