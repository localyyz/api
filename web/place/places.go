package place

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/pressly/chi"
	"github.com/pressly/chi/render"
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

		place, err := data.DB.Place.FindByID(placeID)
		if err != nil {
			render.Respond(w, r, err)
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, "place", place)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func GetPlace(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)
	render.Render(w, r, presenter.NewPlace(ctx, place))
}

// ListNearby returns places and promos based on user's last recorded geolocation
func ListNearby(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)
	cursor := api.NewPage(r)

	query := data.DB.Place.
		Find(db.Cond{"locale_id": user.Etc.LocaleID}).
		Select(
			db.Raw("*"),
			db.Raw(fmt.Sprintf("ST_Distance(geo, st_geographyfromtext('%v'::text)) distance", user.Geo)),
		).
		OrderBy("distance")
	query = cursor.UpdateQueryUpper(query)

	var places []*data.Place
	if err := query.All(&places); err != nil {
		render.Respond(w, r, err)
		return
	}

	presented := presenter.NewPlaceList(ctx, places)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}
}

// ListRecent returns the places with most recent promotions
func ListRecent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)

	var places []*data.Place
	q := data.DB.
		Select(
			db.Raw("distinct on (pl.id) pl.*"),
			db.Raw(fmt.Sprintf("ST_Distance(pl.geo, st_geographyfromtext('%v'::text)) distance", user.Geo)),
		).
		From("places pl").
		LeftJoin("promos pr").
		On("pl.id = pr.place_id").
		Where(db.Cond{
			"pr.status": data.PromoStatusActive,
		}).
		GroupBy("pl.id").
		OrderBy("pl.id").
		Limit(10)

	if err := q.All(&places); err != nil {
		render.Respond(w, r, err)
		return
	}

	presented := presenter.NewPlaceList(ctx, places)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Render(w, r, nil)
	}
}

// Share a place on social media
func Share(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)
	place := ctx.Value("place").(*data.Place)

	payload := &shareWrapper{}
	if err := render.Bind(r, payload); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}
	newShare := &payload.Share
	newShare.UserID = user.ID
	newShare.PlaceID = place.ID
	newShare.Reach = user.Etc.FbFriendCount
	newShare.NetworkShareID = payload.NetworkShareID

	if err := data.DB.Share.Save(newShare); err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Status(r, http.StatusCreated)
	render.Respond(w, r, newShare)
}
