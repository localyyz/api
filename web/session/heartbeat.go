package session

import (
	"net/http"

	"github.com/goware/lg"

	"upper.io/db.v2"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
)

func Heartbeat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)

	var payload struct {
		Longitude float64 `json:"lng,required"`
		Latitude  float64 `json:"lat,required"`
	}
	if err := ws.Bind(r.Body, &payload); err != nil {
		ws.Respond(w, http.StatusBadRequest, err)
		return
	}

	if err := user.SetLocation(payload.Latitude, payload.Longitude); err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	var locale *data.Locale
	cell, err := data.DB.Cell.FindByLatLng(payload.Latitude, payload.Longitude)
	if err != nil {
		if err != db.ErrNoMoreRows {
			ws.Respond(w, http.StatusInternalServerError, err)
			return
		}
		neighbours, err := data.DB.Cell.FindNeighbourByLatLng(payload.Latitude, payload.Longitude)
		if err != nil {
			ws.Respond(w, http.StatusInternalServerError, err)
			return
		}
		for _, n := range neighbours { // assign cell as first one
			cell = n
			break
		}
	}

	if cell != nil {
		locale, err = data.DB.Locale.FindByID(cell.LocaleID)
		if err != nil {
			ws.Respond(w, http.StatusInternalServerError, err)
			return
		}
		// save it into user
		user.Etc.LocaleID = locale.ID
		if err := data.DB.User.Save(user); err != nil {
			ws.Respond(w, http.StatusInternalServerError, err)
			return
		}
		lg.Debugf("user(%d) located at %s", user.ID, locale.Name)
	}
	// NOTE if we didn't find a valid locale, we keep user's previous

	resp := presenter.User{
		User:   user,
		Geo:    user.Geo,
		Locale: locale,
	}
	ws.Respond(w, http.StatusCreated, resp)
}
