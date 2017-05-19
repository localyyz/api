package auth

import (
	"net/http"
	"time"

	"github.com/goware/lg"
	"github.com/pkg/errors"

	"golang.org/x/crypto/bcrypt"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
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

var (
	// Password and Confirmation must match
	ErrPasswordMismatch = errors.New("password mismatch")
	// Minimum length requirement for password
	ErrPasswordLength = errors.New("password must be at least 8 characters long")
	// Unknown encryption error
	ErrEncryptinError = errors.New("internal error")
)

func EmailSignup(w http.ResponseWriter, r *http.Request) {
	var newSignup userSignup
	if err := ws.Bind(r.Body, &newSignup); err != nil {
		ws.Respond(w, http.StatusBadRequest, err)
		return
	}

	if newSignup.Password != newSignup.PasswordConfirm {
		ws.Respond(w, http.StatusBadRequest, ErrPasswordMismatch)
		return
	}

	if len(newSignup.Password) < MinPasswordLength {
		ws.Respond(w, http.StatusBadRequest, ErrPasswordLength)
		return
	}

	// encrypt with bcrypt
	epw, err := bcrypt.GenerateFromPassword([]byte(newSignup.Password), bCryptCost)
	if err != nil {
		// mask the encryption error and return
		ws.Respond(w, http.StatusInternalServerError, ErrEncryptinError)
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
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	// TODO: email verification
	authUser, err := NewAuthUser(newUser)
	if err != nil {
		ws.Respond(w, http.StatusUnauthorized, err)
		return
	}

	ws.Respond(w, http.StatusOK, authUser)
}
