package presenter

import (
	"context"
	"net/http"
	"time"

	"bitbucket.org/moodie-app/moodie-api/data"
)

type User struct {
	*data.User

	Addresses   []*data.UserAddress `json:"addresses"`
	AutoOnboard bool                `json:"autoOnboard"`
}

const (
	MaxOnboardPromptDuration = 120 * time.Second
)

func NewUser(ctx context.Context, user *data.User) *User {
	u := &User{
		User:      user,
		Addresses: make([]*data.UserAddress, 0, 0),
	}

	if dbAddresses, _ := data.DB.UserAddress.FindByUserID(user.ID); dbAddresses != nil {
		u.Addresses = dbAddresses
	}

	u.AutoOnboard = user.Etc.AutoOnboard && time.Since(*user.CreatedAt) < MaxOnboardPromptDuration

	return u
}

func (u *User) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
