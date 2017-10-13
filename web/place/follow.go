package place

import (
	"fmt"
	"net/http"

	"github.com/go-chi/render"

	"upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
)

func ListFollowing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)
	cursor := api.NewPage(r)

	followings, err := data.DB.Following.FindByUserID(user.ID)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	placeIDs := make([]int64, len(followings))
	for i, f := range followings {
		placeIDs[i] = f.PlaceID
	}

	query := data.DB.Place.
		Find(db.Cond{"id": placeIDs}).
		Select(
			db.Raw("*"),
			db.Raw(fmt.Sprintf("ST_Distance(geo, st_geographyfromtext('%v'::text)) distance", user.Geo)),
		).
		OrderBy("distance")
	query = cursor.UpdateQueryUpper(query)

	var places []*data.Place
	if err := query.All(&places); err != nil {
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
