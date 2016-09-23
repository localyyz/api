package web

import (
	"io/ioutil"
	"net/http"
	"os"

	"bitbucket.org/moodie-app/moodie-api/lib/pusher"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
	"bitbucket.org/moodie-app/moodie-api/web/auth"
	"bitbucket.org/moodie-app/moodie-api/web/place"
	"bitbucket.org/moodie-app/moodie-api/web/session"
	"bitbucket.org/moodie-app/moodie-api/web/user"

	"github.com/pkg/errors"
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

	r.Get("/manifest.plist", func(w http.ResponseWriter, r *http.Request) {
		payload := `<?xml version="1.0" encoding="UTF-8"?>
		<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
		<plist version="1.0">
		<dict>
			<key>items</key>
			<array>
				<dict>
					<key>assets</key>
					<array>
						<dict>
							<key>kind</key>
							<string>software-package</string>
							<key>url</key>
							<string>https://api.localyyz.com/Moodie.ipa</string>
						</dict>
						<dict>
							<key>kind</key>
							<string>display-image</string>
							<key>url</key>
							<string>http://localyyz.com/57.png</string>
						</dict>
						<dict>
							<key>kind</key>
							<string>full-size-image</string>
							<key>url</key>
							<string>https://localyyz.com/512.png</string>
						</dict>
					</array>
					<key>metadata</key>
					<dict>
						<key>bundle-identifier</key>
						<string>org.moodie.toronto</string>
						<key>bundle-version</key>
						<string>1.0.0</string>
						<key>kind</key>
						<string>software</string>
						<key>title</key>
						<string>Moodie</string>
					</dict>
				</dict>
			</array>
		</dict>
		</plist>`
		w.Header().Set("Content-Disposition", "attachement; filename=manifest.plist")
		w.Write([]byte(payload))
	})
	r.Get("/Moodie.ipa", func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("/etc/Moodie.ipa")
		if err != nil {
			ws.Respond(w, 404, errors.Wrap(err, "app download error"))
			return
		}
		defer f.Close()
		w.Header().Set("Content-Disposition", "attachement; filename=Moodie.ipa")
		binary, err := ioutil.ReadAll(f)
		if err != nil {
			ws.Respond(w, 404, errors.Wrap(err, "app download error"))
			return
		}
		w.Write(binary)
	})

	r.Post("/login/facebook", auth.FacebookLogin)
	r.Post("/echo", echoPush)

	r.Group(func(r chi.Router) {
		r.Use(session.SessionCtx)

		r.Mount("/session", session.Routes())
		r.Mount("/users", user.Routes())
		r.Mount("/places", place.Routes())
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
