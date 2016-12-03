package session

import (
	"encoding/json"
	"net/http"

	"github.com/goware/lg"

	db "upper.io/db.v2"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
)

type Heartbeat struct {
	Longitude        float64     `json:"longitude"`
	Latitude         float64     `json:"latitude"`
	Speed            int64       `json:"speed"`
	Time             float64     `json:"time"`
	LocationType     string      `json:"location_type"`
	Accuracy         int32       `json:"accuracy"`
	Heading          float64     `json:"heading"`
	Altitude         json.Number `json:"altitude"`
	AltitudeAccuracy json.Number `json:"altitudeAccuracy"`
}

func PostHeartbeat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)

	var payload []*Heartbeat
	if err := ws.BindMany(r.Body, &payload); err != nil {
		ws.Respond(w, http.StatusBadRequest, err)
		return
	}

	// TODO: should sort by timestamp
	// for now, just take and forget
	if len(payload) == 0 {
		ws.Respond(w, http.StatusOK, "")
		return
	}

	newCoord := payload[0]
	// save the user's location as a geohash
	if err := user.SetLocation(newCoord.Latitude, newCoord.Longitude); err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	// find the cell the coord falls under.
	// if we can't find one, find a neighbouring one
	cell, err := data.DB.Cell.FindByLatLng(newCoord.Latitude, newCoord.Longitude)
	if err != nil {
		if err != db.ErrNoMoreRows {
			ws.Respond(w, http.StatusInternalServerError, err)
			return
		}
		neighbours, err := data.DB.Cell.FindNeighbourByLatLng(newCoord.Latitude, newCoord.Longitude)
		if err != nil {
			ws.Respond(w, http.StatusInternalServerError, err)
			return
		}
		for _, n := range neighbours { // assign cell as first one
			cell = n
			break
		}
	}

	var locale *data.Locale
	// if we found a cell, find the neighbourhood
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
