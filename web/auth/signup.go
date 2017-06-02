package auth

import (
	"net/http"
	"time"

	"github.com/goware/lg"
	"github.com/pkg/errors"
	"github.com/pressly/chi/render"

	"golang.org/x/crypto/bcrypt"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/web/api"
)

type userSignup struct {
	FullName        string          `json:"fullName,required"`
	Email           string          `json:"email,required"`
	Password        string          `json:"password,required"`
	PasswordConfirm string          `json:"passwordConfirm,required"`
	Dob             time.Time       `json:"dob,required"`
	Gender          data.UserGender `json:"gender,required"`
}

const (
	MinPasswordLength int = 8
	bCryptCost        int = 10
)

func (u *userSignup) Bind(r *http.Request) error {
	return nil
}

func EmailSignup(w http.ResponseWriter, r *http.Request) {
	newSignup := &userSignup{}
	if err := render.Bind(r, newSignup); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	if newSignup.Password != newSignup.PasswordConfirm {
		render.Respond(w, r, api.WrapErr(api.ErrPasswordMismatch))
		return
	}

	if len(newSignup.Password) < MinPasswordLength {
		render.Respond(w, r, api.WrapErr(api.ErrPasswordLength))
		return
	}

	// encrypt with bcrypt
	epw, err := bcrypt.GenerateFromPassword([]byte(newSignup.Password), bCryptCost)
	if err != nil {
		// mask the encryption error and return
		render.Respond(w, r, api.WrapErr(api.ErrEncryptinError))
		lg.Alert(errors.Wrap(err, "encryption error"))
		return
	}

	newUser := &data.User{
		Username:     newSignup.Email,
		Email:        newSignup.Email,
		Name:         newSignup.FullName,
		PasswordHash: string(epw),
		LastLogInAt:  data.GetTimeUTCPointer(),
		LoggedIn:     true,
	}
	if err := data.DB.User.Save(newUser); err != nil {
		render.Render(w, r, api.WrapErr(err))
		return
	}

	// TODO: email verification
	authUser := NewAuthUser(newUser)
	if err := render.Render(w, r, authUser); err != nil {
		render.Render(w, r, api.WrapErr(err))
	}
}
