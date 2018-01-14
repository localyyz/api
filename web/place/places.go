package place

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/pressly/lg"

	"upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
)

type shareWrapper struct {
	data.Share

	NetworkShareID string      `json:"networkShareId"`
	ID             interface{} `json:"id,omitempty"`
	UserID         interface{} `json:"userId,omitempty"`
	PlaceID        interface{} `json:"userId,omitempty"`
	CreatedAt      interface{} `json:"createdAt,omitempty"`
}

func (s *shareWrapper) Bind(r *http.Request) error {
	return nil
}

func PlaceCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		placeID, err := strconv.ParseInt(chi.URLParam(r, "placeID"), 10, 64)
		if err != nil {
			render.Render(w, r, api.ErrBadID)
			return
		}

		var place *data.Place
		err = data.DB.Place.Find(
			db.Cond{
				"id":     placeID,
				"status": data.PlaceStatusActive,
			},
		).One(&place)
		if err != nil {
			render.Respond(w, r, err)
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, "place", place)
		lg.SetEntryField(ctx, "place_id", place.ID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func GetPlace(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)
	render.Render(w, r, presenter.NewPlace(ctx, place))
}

// Share a place on social media
func Share(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("session.user").(*data.User)
	if !ok {
		return
	}
	place := ctx.Value("place").(*data.Place)

	payload := &shareWrapper{}
	if err := render.Bind(r, payload); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}
	newShare := &payload.Share
	newShare.UserID = user.ID
	newShare.PlaceID = place.ID
	newShare.NetworkShareID = payload.NetworkShareID

	if err := data.DB.Share.Save(newShare); err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Status(r, http.StatusCreated)
	render.Respond(w, r, newShare)
}
