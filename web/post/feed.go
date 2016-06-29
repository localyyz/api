package post

import (
	"net/http"

	"upper.io/db"

	"github.com/goware/lg"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"

	"golang.org/x/net/context"
)

func ListTrendingPost(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	cursor := ws.NewPage(r)
	posts, err := data.DB.Post.GetTrending(cursor)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}
	resp := []*data.PostPresenter{}
	for _, p := range posts {
		u, err := data.DB.User.FindByID(p.UserID)
		if err != nil {
			lg.Warn(err)
			continue
		}
		resp = append(resp, &data.PostPresenter{Post: p, User: u})
	}

	ws.Respond(w, http.StatusOK, resp, cursor.Update(resp))
}

func ListFreshPost(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	cursor := ws.NewPage(r)
	posts, err := data.DB.Post.GetFresh(cursor)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	resp := []*data.PostPresenter{}
	for _, post := range posts {
		user, err := data.DB.User.FindByID(post.UserID)
		if err != nil {
			lg.Warn(err)
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
