package post

import (
	"net/http"

	"upper.io/db.v2"

	"github.com/goware/lg"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
)

func ListFreshPost(w http.ResponseWriter, r *http.Request) {
	place, hasPlaceCtx := r.Context().Value("place").(*data.Place)

	cond := db.Cond{}
	if hasPlaceCtx {
		cond["place_id"] = place.ID
	}

	cursor := ws.NewPage(r)
	posts, err := data.DB.Post.GetFresh(cursor, cond)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	resp := []*data.PostPresenter{}
	for _, post := range posts {
		user, err := data.DB.User.FindByID(post.UserID)
		if err != nil {
			continue
		}
		place, err := data.DB.Place.FindByID(post.PlaceID)
		if err != nil {
			lg.Warn(err)
			continue
		}

		liked, err := data.DB.Like.Find(db.Cond{"user_id": post.UserID, "post_id": post.ID}).Count()
		if err != nil {
			continue
		}

		userContext := &data.UserContext{
			Liked: (liked > 0),
		}

		resp = append(resp, &data.PostPresenter{Post: post, User: user, Place: place, Context: userContext})
	}

	ws.Respond(w, http.StatusOK, resp, cursor.Update(resp))
}
