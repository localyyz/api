package user

import (
	"net/http"
	"strconv"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
	"bitbucket.org/moodie-app/moodie-api/web/utils"

	"github.com/pressly/chi"

	"upper.io/db"

	"golang.org/x/net/context"
)

func MeCtx(next chi.Handler) chi.Handler {
	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		me, ok := ctx.Value("session.user").(*data.User)
		if !ok {
			ws.Respond(w, http.StatusUnauthorized, nil)
			return
		}
		ctx = context.WithValue(ctx, "user", me)
		next.ServeHTTPC(ctx, w, r)
	}
	return chi.HandlerFunc(handler)
}

func UserCtx(next chi.Handler) chi.Handler {
	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		userID, err := strconv.ParseInt(chi.URLParam(ctx, "userID"), 10, 64)
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
		next.ServeHTTPC(ctx, w, r)
	}
	return chi.HandlerFunc(handler)
}

func GetUser(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	user := ctx.Value("user").(*data.User)
	ws.Respond(w, http.StatusOK, user)
}

func GetPointHistory(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	user := ctx.Value("session.user").(*data.User)
	points, err := data.DB.UserPoint.FindByUserID(user.ID)
	if err != nil {
		ws.Respond(w, http.StatusServiceUnavailable, err)
		return
	}

	// no.. i don't want to load posts thru likes..
	type presenter struct {
		*data.UserPoint
		Post   *data.Post `json:"post"`
		Reward uint32     `json:"reward"`
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
			Reward:    10, // TODO: Arbituary.
		}
	}

	ws.Respond(w, http.StatusOK, points)
	return
}

func GetRecentPost(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	user := ctx.Value("user").(*data.User)

	posts, err := data.DB.Post.FindUserRecent(user.ID)
	if err != nil {
		ws.Respond(w, http.StatusServiceUnavailable, err)
		return
	}

	ws.Respond(w, http.StatusOK, posts)
	return
}
