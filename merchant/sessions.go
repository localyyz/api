package merchant

import (
	"context"
	"encoding/json"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
)

func SessionCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		token, claims, err := jwtauth.FromContext(ctx)
		if token == nil || err != nil {
			lg.Alertf("invalid merchant ctx token is nil (%s): %+v", r.URL.Path, err)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		rawPlaceID, ok := claims["place_id"].(json.Number)
		if !ok {
			lg.Error("invalid session merchant, no place_id found")
			return
		}

		placeID, err := rawPlaceID.Int64()
		if err != nil {
			lg.Errorf("invalid session merchant: %+v", err)
			return
		}

		// find a logged in user with the given id
		place, err := data.DB.Place.FindOne(
			db.Cond{"id": placeID},
		)
		if err != nil {
			lg.Errorf("invalid session merchant: %+v", err)
			return
		}

		ctx = context.WithValue(ctx, "place", place)
		ctx = context.WithValue(ctx, "session.place", place)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func MustValidateSessionCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		_, ok := r.Context().Value("session.place").(*data.Place)
		if !ok {
			// no session. return forbidden
			render.Respond(w, r, api.ErrInvalidSession)
			return
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(handler)
}
