package place

import (
	"net/http"

	"upper.io/db.v3"

	"github.com/pressly/chi/render"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
)

func ListFollowing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)

	followings, err := data.DB.Following.FindByUserID(user.ID)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	placeIDs := make([]int64, len(followings))
	for i, f := range followings {
		placeIDs[i] = f.PlaceID
	}

	places, err := data.DB.Place.FindAll(db.Cond{"id": placeIDs})
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	presented := presenter.NewPlaceList(ctx, places)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
		return
	}
}

func FollowPlace(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)
	user := ctx.Value("session.user").(*data.User)

	follow := &data.Following{
		UserID:  user.ID,
		PlaceID: place.ID,
	}
	if err := data.DB.Following.Save(follow); err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Status(r, http.StatusCreated)
	render.Respond(w, r, follow)
}

func UnfollowPlace(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)
	user := ctx.Value("session.user").(*data.User)

	cond := db.Cond{
		"user_id":  user.ID,
		"place_id": place.ID,
	}
	following, err := data.DB.Following.FindOne(cond)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	if err := data.DB.Following.Delete(following); err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Status(r, http.StatusNoContent)
	render.Respond(w, r, "")
}
