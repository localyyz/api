package place

import (
	"net/http"

	db "upper.io/db.v2"

	"github.com/pkg/errors"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
)

func ListFollowing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)

	followings, err := data.DB.Following.FindByUserID(user.ID)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, errors.Wrap(err, "failed to find favorite places"))
		return
	}

	placeIDs := make([]int64, len(followings))
	for i, f := range followings {
		placeIDs[i] = f.PlaceID
	}

	places, err := data.DB.Place.FindAll(db.Cond{"id": placeIDs})
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, errors.Wrap(err, "failed to query favorite places"))
		return
	}

	response := make([]*presenter.Place, len(places))
	for i, pl := range places {
		response[i] = presenter.NewPlace(ctx, pl).WithPromo().WithLocale().WithGeo()
	}

	// TODO: present
	ws.Respond(w, http.StatusOK, response)
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
		ws.Respond(w, http.StatusInternalServerError, errors.Wrap(err, "following place"))
		return
	}

	ws.Respond(w, http.StatusCreated, follow)
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
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	if err := data.DB.Following.Delete(following); err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	ws.Respond(w, http.StatusNoContent, "")
}
