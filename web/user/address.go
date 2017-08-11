package user

import (
	"errors"
	"net/http"

	"github.com/pressly/chi/render"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/web/api"
)

type userAddressRequest struct {
	*data.UserAddress

	userID    interface{} `json:"userId,omitempty"`
	createdAt interface{} `json:"createdAt,omitempty"`
	updatedAt interface{} `json:"createdAt,omitempty"`
	deletedAt interface{} `json:"createdAt,omitempty"`
}

func (u *userAddressRequest) Bind(r *http.Request) error {
	if u.UserAddress == nil {
		return errors.New("address is required")
	}
	return nil
}

func ListAddresses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)

	dbAddresses, err := data.DB.UserAddress.FindByUserID(user.ID)
	if err != nil {
		render.Respond(w, r, err)
		return
	}
	if dbAddresses == nil {
		dbAddresses = make([]*data.UserAddress, 0, 0)
	}
	render.Respond(w, r, dbAddresses)
}

func CreateAddress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)

	payload := &userAddressRequest{}
	if err := render.Bind(r, payload); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	userAddress := payload.UserAddress
	userAddress.UserID = user.ID
	if err := data.DB.Save(userAddress); err != nil {
		render.Respond(w, r, err)
		return
	}
	render.Respond(w, r, userAddress)
}
