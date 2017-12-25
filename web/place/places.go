package place

import (
	"context"
	"fmt"
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

// ListNearby returns places and products based on user's last recorded geolocation
func ListNearby(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("session.user").(*data.User)
	if !ok {
		return
	}

	cursor := api.NewPage(r)

	query := data.DB.Place.
		Find(
			db.Cond{
				"locale_id": user.Etc.LocaleID,
				"status":    data.PlaceStatusActive,
			},
		).
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

// ListRecent returns the places with most recent products
// by product last updated at in descending order
func ListRecent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cursor := api.NewPage(r)

	var orderedPlaces []*data.Place
	query := data.DB.
		Select(
			"pl.id",
			db.Raw("max(pr.updated_at) updated_at"),
		).
		From("places pl").
		LeftJoin("products pr").
		On("pl.id = pr.place_id").
		Where(db.Cond{
			"pl.status": data.PlaceStatusActive,
		}).
		GroupBy("pl.id").
		OrderBy(db.Raw("updated_at DESC NULLS LAST, id DESC"))
	paginator := cursor.UpdateQueryBuilder(query)
	if err := paginator.All(&orderedPlaces); err != nil {
		render.Respond(w, r, err)
		return
	}

	orderedPlaceIDs := make([]int64, len(orderedPlaces))
	for i, pl := range orderedPlaces {
		orderedPlaceIDs[i] = pl.ID
	}

	var places []*data.Place
	err := data.DB.Place.Find(
		db.Cond{"id": orderedPlaceIDs},
	).OrderBy(
		data.MaintainOrder("id", orderedPlaceIDs),
	).All(&places)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	presented := presenter.NewPlaceList(ctx, places)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}
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
