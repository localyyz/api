package me

import (
	"net/http"

	"upper.io/db"

	"bitbucket.org/pxue/api/data"
	"bitbucket.org/pxue/api/lib/ws"

	"golang.org/x/net/context"
)

func GetMe(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	me := ctx.Value("session.user").(*data.User)
	ws.Respond(w, http.StatusOK, me)
}

func GetPoints(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	me := ctx.Value("session.user").(*data.User)
	points, err := data.DB.UserPoint.FindByUserID(me.ID)
	if err != nil {
		ws.Respond(w, http.StatusServiceUnavailable, err)
		return
	}

	// no.. i don't want to load posts thru likes..
	type presenter struct {
		*data.UserPoint
		Post *data.Post
	}

	// get all the posts
	var postID []int64
	for _, p := range points {
		postID = append(postID, p.PostID)
	}
	posts, err := data.DB.Post.FindAll(db.Cond{"post_id": postID})
	if err != nil {
		ws.Respond(w, http.StatusServiceUnavailable, err)
		return
	}
	postsMap := map[int64]*data.Post{}
	for _, p := range posts {
		postsMap[p.ID] = p
	}

	resp := make([]*presenter, len(points))
	for i, p := range points {
		resp[i] = &presenter{
			UserPoint: p,
			Post:      postsMap[p.PostID],
		}
	}

	ws.Respond(w, http.StatusOK, points)
	return
}
