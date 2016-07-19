package post

import (
	"net/http"

	"upper.io/db"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
)

func ListPostLike(w http.ResponseWriter, r *http.Request) {
	post := r.Context().Value("post").(*data.Post)
	likes, err := data.DB.Like.FindByPostID(post.ID)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}
	ws.Respond(w, 200, likes)
}

func LikePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)
	post := ctx.Value("post").(*data.Post)

	newLike := &data.Like{
		PostID: post.ID,
		UserID: user.ID,
	}
	if err := data.DB.Like.Save(newLike); err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}
	ws.Respond(w, http.StatusCreated, newLike)
}

func UnlikePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)
	post := ctx.Value("post").(*data.Post)

	var like *data.Like
	err := data.DB.Like.Find(db.Cond{"user_id": user.ID, "post_id": post.ID}).One(&like)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	if err := data.DB.Like.Delete(like); err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}
	ws.Respond(w, http.StatusNoContent, nil)
}
