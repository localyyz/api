package session

import (
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/maps"
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

	locale, err := maps.GetLocale(ctx, &user.Geo)
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

	resp := data.LocateUser{
		User:   user,
		Geo:    user.Geo,
		Locale: locale,
	}
	ws.Respond(w, http.StatusCreated, resp)
}
