package web

import (
	"io/ioutil"
	"net/http"
	"os"

	"bitbucket.org/moodie-app/moodie-api/lib/ws"
	"bitbucket.org/moodie-app/moodie-api/web/auth"
	"bitbucket.org/moodie-app/moodie-api/web/place"
	"bitbucket.org/moodie-app/moodie-api/web/post"
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

	r.Get("/app", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "itms-services://?action=download-manifest&url=http://127.0.0.1:5331/manifest.plist", 301)
	})
	r.Get("/manifest.plist", func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("/data/app/manifest.plist")
		if err != nil {
			ws.Respond(w, 404, errors.Wrap(err, "app download error"))
			return
		}
		defer f.Close()
		manifest, err := ioutil.ReadAll(f)
		if err != nil {
			ws.Respond(w, 404, errors.Wrap(err, "app download error"))
			return
		}
		w.Write(manifest)
	})
	r.Get("/Moodie.ipa", func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("/data/app/Moodie.ipa")
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

	r.Group(func(r chi.Router) {
		r.Use(session.SessionCtx)

		r.Mount("/session", session.Routes())
		r.Mount("/users", user.Routes())
		r.Mount("/places", place.Routes())
		r.Mount("/posts", post.Routes())
	})

	return r
}
