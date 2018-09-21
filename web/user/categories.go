package user

import (
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/go-chi/render"
)

type categoryRequest struct {
	CategoryIDs []int64 `json:"categories"`
}

func (*categoryRequest) Bind(r *http.Request) error {
	return nil
}

func UpdateCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)

	var catRequest categoryRequest
	if err := render.Bind(r, &catRequest); err != nil {
		render.Respond(w, r, err)
		return
	}
	user.Etc.CategoryIDs = catRequest.CategoryIDs
	if err := data.DB.User.Save(user); err != nil {
		render.Respond(w, r, err)
		return
	}
	render.Respond(w, r, user)
}
