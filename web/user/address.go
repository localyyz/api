package user

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/web/api"
)

type userAddressRequest struct {
	*data.UserAddress

	ID        interface{} `json:"id,omitempty"`
	UserID    interface{} `json:"userId,omitempty"`
	CreatedAt interface{} `json:"createdAt,omitempty"`
	UpdatedAt interface{} `json:"updatedAt,omitempty"`
	DeletedAt interface{} `json:"deletedAt,omitempty"`
}

func (u *userAddressRequest) Bind(r *http.Request) error {
	if u.UserAddress == nil {
		return errors.New("address is required")
	}
	return nil
}

func AddressCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		addressID, err := strconv.ParseInt(chi.URLParam(r, "addressID"), 10, 64)
		if err != nil {
			render.Render(w, r, api.ErrBadID)
			return
		}
		ctx := r.Context()
		user := ctx.Value("session.user").(*data.User)
		var address *data.UserAddress
		err = data.DB.UserAddress.Find(
			db.Cond{
				"id":      addressID,
				"user_id": user.ID,
			},
		).One(&address)
		if err != nil {
			render.Respond(w, r, err)
			return
		}
		ctx = context.WithValue(ctx, "address", address)
		lg.SetEntryField(ctx, "address_id", address.ID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func GetAddress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	address := ctx.Value("address").(*data.UserAddress)
	render.Respond(w, r, address)
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
	if err := data.DB.UserAddress.Save(userAddress); err != nil {
		render.Respond(w, r, err)
		return
	}
	render.Respond(w, r, userAddress)
}

func UpdateAddress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	address := ctx.Value("address").(*data.UserAddress)

	payload := &userAddressRequest{UserAddress: address}
	if err := render.Bind(r, payload); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	if err := data.DB.UserAddress.Save(address); err != nil {
		render.Respond(w, r, err)
		return
	}
	render.Respond(w, r, address)
}

func RemoveAddress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	address := ctx.Value("address").(*data.UserAddress)

	if err := data.DB.UserAddress.Delete(address); err != nil {
		render.Respond(w, r, err)
		return
	}
	render.Status(r, http.StatusNoContent)
	render.Respond(w, r, "")
}
