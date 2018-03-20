package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/pressly/lg"

	"golang.org/x/crypto/bcrypt"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/web/api"
)

type userSignup struct {
	FullName        string          `json:"fullName,required"`
	Email           string          `json:"email,required"`
	Password        string          `json:"password,required"`
	PasswordConfirm string          `json:"passwordConfirm,required"`
	Dob             time.Time       `json:"dob"`
	Gender          data.UserGender `json:"gender"`
	InviteCode      string          `json:"inviteCode"`
}

const (
	MinPasswordLength int = 8
	bCryptCost        int = 10
)

func (u *userSignup) Bind(r *http.Request) error {
	return nil
}

// Signup via website
func RegisterSignup(w http.ResponseWriter, r *http.Request) {
	var newRegister struct {
		Email string `json:"email"`
	}
	err := json.NewDecoder(r.Body).Decode(&newRegister)
	if err != nil {
		render.Respond(w, r, api.ErrInvalidRequest(err))
		return
	}
	if len(newRegister.Email) == 0 {
		return
	}

	newUser := &data.User{
		Username:     newRegister.Email,
		Email:        newRegister.Email,
		EmailStatus:  data.EmailStatusUnconfirmed,
		Name:         newRegister.Email,
		Network:      "website",
		PasswordHash: "",
		LastLogInAt:  data.GetTimeUTCPointer(),
		LoggedIn:     false,
	}
	if err := data.DB.User.Save(newUser); err != nil {
		render.Respond(w, r, err)
		return
	}
	return
}

func EmailSignup(w http.ResponseWriter, r *http.Request) {
	newSignup := &userSignup{}
	if err := render.Bind(r, newSignup); err != nil {
		render.Respond(w, r, api.ErrInvalidRequest(err))
		return
	}

	if newSignup.Password != newSignup.PasswordConfirm {
		render.Respond(w, r, api.ErrPasswordMismatch)
		return
	}

	if len(newSignup.Password) < MinPasswordLength {
		render.Respond(w, r, api.ErrPasswordLength)
		return
	}

	// encrypt with bcrypt
	epw, err := bcrypt.GenerateFromPassword([]byte(newSignup.Password), bCryptCost)
	if err != nil {
		lg.Warn(err)
		// mask the encryption error and return
		render.Respond(w, r, api.ErrEncryptinError)
		return
	}

	newUser := &data.User{
		Username:     newSignup.Email,
		Email:        newSignup.Email,
		Name:         newSignup.FullName,
		Network:      "email",
		PasswordHash: string(epw),
		LastLogInAt:  data.GetTimeUTCPointer(),
		LoggedIn:     true,
	}

	// check if invite code exists
	if newSignup.InviteCode != "" {
		invitor, err := data.DB.User.FindByInviteCode(newSignup.InviteCode)
		if err != nil {
			lg.Warnf("invitor with code %s lookup error: %v", newSignup.InviteCode, err)
			return
		}
		lg.Info("new user invited by: %d", invitor.ID)
		newUser.Etc = data.UserEtc{
			InvitedBy: invitor.ID,
		}
		// TODO return invalid code error?
	}

	if err := data.DB.User.Save(newUser); err != nil {
		render.Respond(w, r, err)
		return
	}

	// TODO: email verification
	ctx := r.Context()
	authUser := NewAuthUser(ctx, newUser)
	if err := render.Render(w, r, authUser); err != nil {
		render.Respond(w, r, err)
	}
}
