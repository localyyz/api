package session

import (
	"encoding/json"
	"net/http"

	db "upper.io/db.v3"

	"github.com/go-chi/render"
	"github.com/pressly/lg"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
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

func (*Heartbeat) Bind(r *http.Request) error {
	return nil
}

func PostHeartbeat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)

	payload := &Heartbeat{}
	if err := render.Bind(r, payload); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	newCoord := payload
	// save the user's location as a geohash
	if err := user.SetLocation(newCoord.Latitude, newCoord.Longitude); err != nil {
		render.Respond(w, r, err)
		return
	}

	locale, err := data.DB.Locale.FromLatLng(newCoord.Latitude, newCoord.Longitude)
	if err != nil && err != db.ErrNoMoreRows {
		render.Respond(w, r, err)
		return
	}

	presented := presenter.NewUser(ctx, user)
	if locale != nil {
		user.Etc.LocaleID = locale.ID
		if err := data.DB.User.Save(user); err != nil {
			render.Respond(w, r, err)
			return
		}

		lg.Infof("user(%d) located at %s", user.ID, locale.Name)
		// NOTE if we didn't find a valid locale, we keep user's previous
		presented.Locale = locale
	}

	// save location history
	ul := &data.UserLocation{
		UserID: user.ID,
		Geo:    user.Geo,
	}
	data.DB.UserLocation.Save(ul)

	render.Render(w, r, presented)
}
