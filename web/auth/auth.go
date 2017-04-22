package auth

import (
	"errors"
	"net/http"

	"github.com/goware/jwtauth"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/token"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
)

// Authenticated user with jwt embed
type AuthUser struct {
	*data.User
	JWT string `json:"jwt"`
}

// AuthUser wraps a user with JWT token
func NewAuthUser(user *data.User) (*AuthUser, error) {
	token, err := token.Encode(jwtauth.Claims{"user_id": user.ID})
	if err != nil {
		return nil, err
	}
	return &AuthUser{User: user, JWT: token.Raw}, nil
}

// FacebookLogin handles both first-time login (signup) and repeated-logins from a social network
// User is already authenticated by the frontend with network of their choice
//  Backend stores the token and async grab the user data
func FacebookLogin(w http.ResponseWriter, r *http.Request) {
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

	authUser, err := NewAuthUser(user)
	if err != nil {
		ws.Respond(w, http.StatusUnauthorized, err)
		return
	}

	ws.Respond(w, http.StatusOK, authUser)
}
