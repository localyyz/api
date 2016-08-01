package session

import (
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
)

func Heartbeat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)

	if err := ws.Bind(r.Body, &user.Geo); err != nil {
		ws.Respond(w, http.StatusBadRequest, err)
		return
	}
	if err := data.DB.User.Save(user); err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	locale, err := data.GetLocale(ctx, &user.Geo)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	resp := data.LocateUser{
		User:   user,
		Locale: locale,
	}
	ws.Respond(w, http.StatusCreated, resp)
}
