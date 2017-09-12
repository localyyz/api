package presenter

import (
	"context"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/goware/geotools"
)

type User struct {
	*data.User

	Addresses []*data.UserAddress `json:"addresses"`

	Geo       geotools.Point `json:"geo"`
	Locale    *data.Locale   `json:"locale"`
	InviteURL string         `json:"inviteUrl"`
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
	u.Geo = u.User.Geo
	return nil
}