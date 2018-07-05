package presenter

import (
	"context"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
)

type User struct {
	*data.User

	Addresses []*data.UserAddress `json:"addresses"`
}

func NewUser(ctx context.Context, user *data.User) *User {
	u := &User{
		User:      user,
		Addresses: make([]*data.UserAddress, 0, 0),
	}

	if dbAddresses, _ := data.DB.UserAddress.FindByUserID(user.ID); dbAddresses != nil {
		u.Addresses = dbAddresses
	}

	return u
}

func (u *User) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
