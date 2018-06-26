package auth

import (
	"context"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	db "upper.io/db.v3"

	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/token"
	"bitbucket.org/moodie-app/moodie-api/web/api"
)

// Authenticated user with jwt embed
type AuthUser struct {
	*data.User
	JWT string `json:"jwt"`
}

type emailLogin struct {
	Email    string `json:"email,required"`
	Password string `json:"password,required"`
}

type fbLogin struct {
	Token      string `json:"token,required"`
	InviteCode string `json:"inviteCode"`
}

var (
	// if the user's password hash is empty, use this
	// hash to mask the fact
	timingHash []byte = []byte("$2a$10$4Kys.PIxpCIoUmlcY6D7QOTuMPgk27lpmV74OWCWfqjwnG/JN4kcu")
)

// Must be a logged in and signed up user
func SessionCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		u, ok := r.Context().Value("session.user").(*data.User)
		if !ok || u.Network == "shadow" {
			// no session. return forbidden
			render.Respond(w, r, api.ErrInvalidSession)
			return
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(handler)
}

func DeviceCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		_, ok := r.Context().Value("session.user").(*data.User)
		if !ok {
			// no session. return forbidden
			render.Respond(w, r, api.ErrInvalidSession)
			return
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(handler)
}

// AuthUser wraps a user with JWT token
func NewAuthUser(ctx context.Context, user *data.User) *AuthUser {
	return &AuthUser{User: user}
}

func (u *AuthUser) Render(w http.ResponseWriter, r *http.Request) error {
	token, err := token.Encode(jwtauth.Claims{"user_id": u.ID})
	if err != nil {
		return err
	}
	u.JWT = token.Raw
	return nil
}

func (l *emailLogin) Bind(r *http.Request) error {
	if len(l.Password) < MinPasswordLength {
		return api.ErrPasswordLength
	}
	return nil
}

func (l *fbLogin) Bind(r *http.Request) error {
	return nil
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
	payload := &emailLogin{}
	if err := render.Bind(r, payload); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	user, err := data.DB.User.FindByUsername(payload.Email)
	if err != nil {
		if err == db.ErrNoMoreRows {
			render.Respond(w, r, api.ErrInvalidLogin)
			return
		}
		render.Respond(w, r, err)
		return
	}

	if !verifyPassword(user.PasswordHash, payload.Password) {
		render.Respond(w, r, api.ErrInvalidLogin)
		return
	}

	ctx := r.Context()
	authUser := NewAuthUser(ctx, user)
	if err := render.Render(w, r, authUser); err != nil {
		render.Respond(w, r, err)
	}
}

// FacebookLogin handles both first-time login (signup) and repeated-logins from a social network
// User is already authenticated by the frontend with network of their choice
//  Backend stores the token and async grab the user data
func FacebookLogin(w http.ResponseWriter, r *http.Request) {
	payload := &fbLogin{}
	if err := render.Bind(r, payload); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	// inspect the token for userId and expiry
	user, err := connect.FB.Login(payload.Token, payload.InviteCode)
	if err != nil {
		if err == connect.ErrTokenExpired {
			render.Status(r, http.StatusUnauthorized)
			render.Respond(w, r, connect.ErrTokenExpired)
			return
		}
		render.Status(r, http.StatusServiceUnavailable)
		render.Respond(w, r, err)
		return
	}

	// 1. check if session device user is set
	// 2. check if this is a new user event
	// 3. merge two users
	if u, ok := r.Context().Value("session.user").(*data.User); ok && u.Network == "shadow" && user.ID == 0 {
		user.ID = u.ID
		user.DeviceToken = &u.Username
	}

	if err := data.DB.User.Save(user); err != nil {
		render.Respond(w, r, err)
		return
	}

	ctx := r.Context()
	authUser := NewAuthUser(ctx, user)
	if err := render.Render(w, r, authUser); err != nil {
		render.Respond(w, r, err)
	}
}
