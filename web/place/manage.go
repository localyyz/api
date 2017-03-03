package place

import (
	"net/http"

	db "upper.io/db.v3"

	"github.com/pkg/errors"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
)

func ListManagable(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("session.user").(*data.User)

	access, err := data.DB.UserAccess.EditorAccess(user.ID)
	if err != nil {
		e := errors.Wrap(err, "unable to find access")
		ws.Respond(w, http.StatusInternalServerError, e)
		return
	}

	// find the places the user can manage
	var placeIDs []int64
	for _, a := range access {
		placeIDs = append(placeIDs, a.PlaceID)
	}

	var places []*data.Place
	err = data.DB.Place.Find(db.Cond{"id": placeIDs}).All(&places)
	if err != nil {
		e := errors.Wrap(err, "unable to find places")
		ws.Respond(w, http.StatusInternalServerError, e)
		return
	}

	ws.Respond(w, http.StatusOK, places)
}
