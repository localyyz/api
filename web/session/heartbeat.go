package session

import (
	"encoding/json"
	"net/http"

	"upper.io/db.v3"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
	"github.com/goware/lg"

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

	latlng := s2.LatLngFromDegrees(newCoord.Latitude, newCoord.Longitude)
	origin := s2.CellIDFromLatLng(latlng).Parent(15) // 16 for more detail?
	// Find the reach of cells
	cond := db.Cond{
		"cell_id >=": int(origin.RangeMin()),
		"cell_id <=": int(origin.RangeMax()),
	}
	cells, err := data.DB.Cell.FindAll(cond)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	// Find the minimum distance cell
	min := s1.InfAngle()
	var localeID int64
	for _, c := range cells {
		cell := s2.CellID(c.CellID)
		d := latlng.Distance(cell.LatLng())
		if d < min {
			min = d
			localeID = c.LocaleID
		}
	}

	resp := presenter.User{
		User: user,
		Geo:  user.Geo,
	}
	if localeID != 0 {
		user.Etc.LocaleID = localeID
		if err := data.DB.User.Save(user); err != nil {
			ws.Respond(w, http.StatusInternalServerError, err)
			return
		}

		locale, err := data.DB.Locale.FindByID(localeID)
		if err != nil {
			ws.Respond(w, http.StatusInternalServerError, err)
			return
		}
		lg.Infof("user(%d) located at %s", user.ID, locale.Name)
		// NOTE if we didn't find a valid locale, we keep user's previous
		resp.Locale = locale
	}

	// save location history
	ul := &data.UserLocation{
		UserID: user.ID,
		Geo:    user.Geo,
	}
	data.DB.UserLocation.Save(ul)

	ws.Respond(w, http.StatusCreated, resp)
}
