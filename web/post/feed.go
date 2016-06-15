package post

import (
	"net/http"

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
