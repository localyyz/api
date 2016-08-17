package place

import (
	"fmt"
	"net/http"

	db "upper.io/db.v2"

	"github.com/pkg/errors"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
)

// ListTrendingNearby returns nearby posts from nearby places ordered by score
func ListTrending(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)

	q := data.DB.Select("pl.id").
		From("places pl").
		Join("posts p").
		On("pl.id = p.place_id").
		GroupBy("pl.id").
		OrderBy(db.Raw(fmt.Sprintf("sum(CASE WHEN pl.locale_id = %d THEN 2^32 ELSE p.score END) DESC", user.Etc.LocaleID))).
		Limit(15)

	var places []*data.Place
	if err := q.All(&places); err != nil {
		ws.Respond(w, http.StatusInternalServerError, errors.Wrap(err, "trending places"))
		return
	}

	resp := struct {
		Nearby   []*data.PlaceWithPost `json:"nearby"`
		Promoted []*data.PlaceWithPost `json:"promoted"`
	}{
		Nearby:   []*data.PlaceWithPost{},
		Promoted: []*data.PlaceWithPost{},
	}

	for _, place := range places {
		var posts []*data.Post
		err := data.DB.Post.
			Find(db.Cond{"place_id": place.ID}).
			OrderBy("-score").
			Limit(5).
			All(&posts)
		if err != nil {
			continue
		}
		postPresenters := make([]*data.PostPresenter, len(posts))
		for i, p := range posts {
			user, err := data.DB.User.FindByID(p.UserID)
			if err != nil {
				continue
			}
			liked, err := data.DB.Like.Find(db.Cond{"user_id": user.ID, "post_id": p.ID}).Count()
			if err != nil {
				continue
			}
			postPresenters[i] = &data.PostPresenter{Post: p, User: user, Context: &data.UserContext{Liked: (liked > 0)}}
		}
		place, err = data.DB.Place.FindByID(place.ID)
		if err != nil {
			continue
		}
		placePresenter := &data.PlaceWithPost{Place: place, Posts: postPresenters}
		if place.LocaleID == user.Etc.LocaleID {
			resp.Nearby = append(resp.Nearby, placePresenter)
		} else {
			resp.Promoted = append(resp.Promoted, placePresenter)
		}
	}

	ws.Respond(w, http.StatusOK, resp)
}
