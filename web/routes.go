package web

import (
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/lib/pusher"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
	"bitbucket.org/moodie-app/moodie-api/web/auth"
	"bitbucket.org/moodie-app/moodie-api/web/place"
	"bitbucket.org/moodie-app/moodie-api/web/promo"
	"bitbucket.org/moodie-app/moodie-api/web/session"
	"bitbucket.org/moodie-app/moodie-api/web/user"

	"github.com/pressly/chi"
	"github.com/pressly/chi/middleware"
)

func New() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.NoCache)
	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`¯\_(ツ)_/¯`))
	})

	// App download related handlers
	r.Get("/manifest.plist", Manifest)
	r.Get("/Moodie.ipa", MoodieApp)

	r.Post("/login/facebook", auth.FacebookLogin)
	r.Post("/echo", echoPush)

	r.Group(func(r chi.Router) {
		r.Use(session.SessionCtx)

		r.Mount("/session", session.Routes())
		r.Mount("/users", user.Routes())
		r.Mount("/places", place.Routes())
		r.Mount("/promos", promo.Routes())
	})

	return r
}

// test function: echo push to apns
func echoPush(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		DeviceToken string `json:"deviceToken"`
		Payload     string `json:"payload"`
	}
	if err := ws.Bind(r.Body, &payload); err != nil {
		ws.Respond(w, http.StatusBadRequest, err)
		return
	}
	err := pusher.Push(payload.DeviceToken, []byte(payload.Payload))
	if err != nil {
		ws.Respond(w, http.StatusBadRequest, err)
		return
	}

	return
}
