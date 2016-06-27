package post

import (
	"net/http"

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
	for _, p := range posts {
		u, err := data.DB.User.FindByID(p.UserID)
		if err != nil {
			lg.Warn(err)
			continue
		}
		l, err := data.DB.Place.FindByID(p.PlaceID)
		if err != nil {
			lg.Warn(err)
			continue
		}
		resp = append(resp, &data.PostPresenter{Post: p, User: u, Place: l})
	}

	ws.Respond(w, http.StatusOK, resp, cursor.Update(resp))
}
