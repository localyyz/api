package user

import (
	"context"
	"net/http"
	"strconv"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
	"bitbucket.org/moodie-app/moodie-api/web/utils"

	"github.com/pressly/chi"

	"upper.io/db"
)

func MeCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		me, ok := ctx.Value("session.user").(*data.User)
		if !ok {
			ws.Respond(w, http.StatusUnauthorized, nil)
			return
		}
		ctx = context.WithValue(ctx, "user", me)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func UserCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
		if err != nil {
			ws.Respond(w, http.StatusBadRequest, utils.ErrBadID)
			return
		}

		user, err := data.DB.User.FindByID(userID)
		if err != nil {
			ws.Respond(w, http.StatusInternalServerError, err)
			return
		}
		ctx = context.WithValue(ctx, "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*data.User)
	ws.Respond(w, http.StatusOK, user)
}

func GetPointHistory(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("session.user").(*data.User)
	points, err := data.DB.UserPoint.FindByUserID(user.ID)
	if err != nil {
		ws.Respond(w, http.StatusServiceUnavailable, err)
		return
	}

	// get all the posts
	var postID []int64
	for _, p := range points {
		postID = append(postID, p.PostID)
	}
	posts, err := data.DB.Post.FindAll(db.Cond{"id": postID})
	if err != nil {
		ws.Respond(w, http.StatusServiceUnavailable, err)
		return
	}
	postsMap := map[int64]*data.Post{}
	for _, p := range posts {
		postsMap[p.ID] = p
	}

	var placeID []int64
	for _, p := range points {
		placeID = append(placeID, p.PlaceID)
	}
	places, err := data.DB.Place.FindAll(db.Cond{"id": placeID})
	if err != nil {
		ws.Respond(w, http.StatusServiceUnavailable, err)
		return
	}
	placesMap := map[int64]*data.Place{}
	for _, p := range places {
		placesMap[p.ID] = p
	}

	resp := make([]*data.UserPointPresenter, len(points))
	for i, p := range points {
		resp[i] = &data.UserPointPresenter{
			UserPoint: p,
			Post:      postsMap[p.PostID],
			Place:     placesMap[p.PlaceID],
			Reward:    10, // TODO: Arbituary.
		}
	}

	ws.Respond(w, http.StatusOK, resp)
	return
}

func GetRecentPost(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*data.User)

	posts, err := data.DB.Post.FindUserRecent(user.ID)
	if err != nil {
		ws.Respond(w, http.StatusServiceUnavailable, err)
		return
	}

	ws.Respond(w, http.StatusOK, posts)
	return
}
