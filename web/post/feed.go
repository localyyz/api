package post

import (
	"context"
	"net/http"

	"bitbucket.org/pxue/api/data"
	"bitbucket.org/pxue/api/lib/ws"
)

func ListTrendingPost(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	cursor := ws.NewPage(r)
	posts, err := data.DB.Post.GetTrending(cursor)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}
	ws.Respond(w, http.StatusOK, posts, cursor.Update(posts))
	return
}

func ListFreshPost(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	cursor := ws.NewPage(r)
	posts, err := data.DB.Post.GetFresh(cursor)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}
	ws.Respond(w, http.StatusOK, posts, cursor.Update(posts))
	return
}
