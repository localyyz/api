package place

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/pkg/errors"
	"github.com/pressly/chi"
)

func PlaceCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		placeID, err := strconv.ParseInt(chi.URLParam(r, "placeID"), 10, 64)
		if err != nil {
			ws.Respond(w, http.StatusBadRequest, api.ErrBadID)
			return
		}

		place, err := data.DB.Place.FindByID(placeID)
		if err != nil {
			ws.Respond(w, http.StatusInternalServerError, err)
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, "place", place)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

// getTrending returns most popular places ordered by aggregated score
func getTrending(user *data.User) ([]*data.Place, error) {
	var places []*data.Place
	q := data.DB.
		Select(
			db.Raw("pl.*"),
			db.Raw(fmt.Sprintf("ST_Distance(pl.geo, st_geographyfromtext('%v'::text)) distance", user.Geo)),
		).
		From("places pl").
		LeftJoin("claims cl").
		On("pl.id = cl.place_id").
		GroupBy("pl.id").
		OrderBy(db.Raw("count(cl) DESC NULLS LAST")).
		Limit(10)
	if err := q.All(&places); err != nil {
		return nil, errors.Wrap(err, "trending places")
	}

	return places, nil
}

func ListPlaces(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUser := ctx.Value("session.user").(*data.User)

	var places []*data.Place
	// if not admin, return
	if !currentUser.IsAdmin {
		ws.Respond(w, http.StatusOK, places)
		return
	}

	var err error
	places, err = data.DB.Place.FindAll(db.Cond{})
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}
	ws.Respond(w, http.StatusOK, places)
}

func GetPlace(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)
	ws.Respond(w, http.StatusOK, (presenter.NewPlace(ctx, place)).WithGeo().WithLocale())
}

// Nearby returns places and promos based on user's last recorded geolocation
func Nearby(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)
	cursor := ws.NewPage(r)

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
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	var presented []*presenter.Place
	for _, pl := range places {
		// TODO: +1 here
		p := presenter.NewPlace(ctx, pl).WithPromo().WithFollowing()
		presented = append(presented, p.WithGeo())
	}

	ws.Respond(w, http.StatusOK, presented, cursor.Update(places))
}

// Recent places returns the most recently created promotions
func Recent(w http.ResponseWriter, r *http.Request) {
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
			"pr.start_at <=": time.Now().UTC(),
			"pr.end_at >":    time.Now().UTC(),
			"pr.status":      data.PromoStatusActive,
		}).
		GroupBy("pl.id", "pr.start_at").
		OrderBy("pl.id", "-pr.start_at").
		Limit(10)

	if err := q.All(&places); err != nil {
		ws.Respond(w, http.StatusInternalServerError, errors.Wrap(err, "recent promotions"))
		return
	}

	var presented []*presenter.Place
	for _, pl := range places {
		// TODO: +1 here
		p := presenter.NewPlace(ctx, pl).WithPromo().WithFollowing()
		presented = append(presented, p.WithGeo())
	}

	ws.Respond(w, http.StatusOK, presented)
}

// Share a place on social media
func Share(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)
	place := ctx.Value("place").(*data.Place)

	var shareWrapper struct {
		data.Share

		NetworkShareID string      `json:"networkShareId"`
		ID             interface{} `json:"id,omitempty"`
		UserID         interface{} `json:"userId,omitempty"`
		PlaceID        interface{} `json:"userId,omitempty"`
		CreatedAt      interface{} `json:"createdAt,omitempty"`
	}
	if err := ws.Bind(r.Body, &shareWrapper); err != nil {
		ws.Respond(w, http.StatusBadRequest, err)
		return
	}
	newShare := &shareWrapper.Share
	newShare.UserID = user.ID
	newShare.PlaceID = place.ID
	newShare.Reach = user.Etc.FbFriendCount
	newShare.NetworkShareID = shareWrapper.NetworkShareID

	if err := data.DB.Share.Save(newShare); err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	ws.Respond(w, http.StatusOK, newShare)
}
