package place

import (
	"net/http"

	"upper.io/db.v2"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
)

// ListTrendingNearby returns nearby posts from nearby places ordered by score
func ListTrending(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)
	cursor := ws.NewPage(r)

	q := data.DB.Place.Find(db.Cond{"locale_id": user.Etc.LocaleID})
	q = cursor.UpdateQueryUpper(q)

	var placeIDs []int64
	nearbyMap := map[int64]*data.Place{}
	var place *data.Place
	for q.Next(&place) {
		placeIDs = append(placeIDs, place.ID)
		nearbyMap[place.ID] = place
	}
	if err := q.Err(); err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
	}

	// order list of places by post score
	var byScore []int64
	err := data.DB.Post.
		Find(db.Cond{"place_id": placeIDs}).
		Select("place_id").
		Group("place_id").
		OrderBy(db.Raw("-SUM(score)")).
		All(&byScore)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	resp := struct {
		Nearby   []*data.PlaceWithPost `json:"nearby"`
		Promoted []*data.PlaceWithPost `json:"promoted"`
	}{
		Nearby:   []*data.PlaceWithPost{},
		Promoted: []*data.PlaceWithPost{},
	}

	for pID, place := range nearbyMap {
		var posts []*data.Post
		err := data.DB.Post.
			Find(db.Cond{"place_id": pID}).
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

			liked, err := data.DB.Like.Find(db.Cond{"user_id": p.UserID, "post_id": p.ID}).Count()
			if err != nil {
				continue
			}

			postPresenters[i] = &data.PostPresenter{Post: p, User: user, Context: &data.UserContext{Liked: (liked > 0)}}
		}
		resp.Nearby = append(resp.Nearby, &data.PlaceWithPost{Place: place, Posts: postPresenters})
	}

	ws.Respond(w, http.StatusOK, resp, cursor.Update(resp))
}
