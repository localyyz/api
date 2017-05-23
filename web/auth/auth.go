package auth

import (
	"errors"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	db "upper.io/db.v3"

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

var (
	// if the user's password hash is empty, use this
	// hash to mask the fact
	timingHash []byte = []byte("$2a$10$4Kys.PIxpCIoUmlcY6D7QOTuMPgk27lpmV74OWCWfqjwnG/JN4kcu")
)

var (
	ErrInvalidLogin = errors.New("invalid login credentials, check username and/or password")
)

// AuthUser wraps a user with JWT token
func NewAuthUser(user *data.User) (*AuthUser, error) {
	token, err := token.Encode(jwtauth.Claims{"user_id": user.ID})
	if err != nil {
		return nil, err
	}
	return &AuthUser{User: user, JWT: token.Raw}, nil
}

// bcrypt compare hash with given password
func verifyPassword(hash, password string) bool {

	// incase either hash or password is empty, compare
	// something and return false to mask the timing
	if len(hash) == 0 || len(password) == 0 {
		bcrypt.CompareHashAndPassword(timingHash, []byte(password))
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func EmailLogin(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email    string `json:"email,required"`
		Password string `json:"password,required"`
	}
	if err := ws.Bind(r.Body, &payload); err != nil {
		ws.Respond(w, http.StatusBadRequest, err)
		return
	}

	if len(payload.Password) < MinPasswordLength {
		ws.Respond(w, http.StatusBadRequest, ErrPasswordLength)
		return
	}

	user, err := data.DB.User.FindByUsername(payload.Email)
	if err != nil {
		if err == db.ErrNoMoreRows {
			ws.Respond(w, http.StatusUnauthorized, ErrInvalidLogin)
			return
		}
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	if !verifyPassword(user.PasswordHash, payload.Password) {
		ws.Respond(w, http.StatusUnauthorized, ErrInvalidLogin)
		return
	}

	authUser, err := NewAuthUser(user)
	if err != nil {
		ws.Respond(w, http.StatusUnauthorized, err)
		return
	}

	ws.Respond(w, http.StatusOK, authUser)
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
