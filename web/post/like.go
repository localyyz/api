package post

import (
	"net/http"
	"strconv"

	"bitbucket.org/pxue/api/data"
	"bitbucket.org/pxue/api/lib/ws"
	"bitbucket.org/pxue/api/web/utils"

	"github.com/pressly/chi"

	"golang.org/x/net/context"
)

func LikeCtx(next chi.Handler) chi.Handler {
	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		likeID, err := strconv.ParseInt(chi.URLParam(ctx, "likeID"), 10, 64)
		if err != nil {
			ws.Respond(w, http.StatusBadRequest, utils.ErrBadID)
			return
		}

		like, err := data.DB.Like.FindByID(likeID)
		if err != nil {
			ws.Respond(w, http.StatusInternalServerError, err)
			return
		}
		ctx = context.WithValue(ctx, "like", like)
		next.ServeHTTPC(ctx, w, r)
	}
	return chi.HandlerFunc(handler)
}

func ListPostLike(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	post := ctx.Value("post").(*data.Post)
	likes, err := data.DB.Like.FindByPostID(post.ID)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}
	ws.Respond(w, 200, likes)
}

func GetLike(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	like := ctx.Value("like").(*data.Like)
	ws.Respond(w, 200, like)
}

func LikePost(ctx context.Context, w http.ResponseWriter, r *http.Request) {
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

func UnlikePost(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	user := ctx.Value("session.user").(*data.User)
	like := ctx.Value("like").(*data.Like)
	if user.ID != like.UserID {
		ws.Respond(w, http.StatusBadRequest, utils.ErrBadAction)
		return
	}
	if err := data.DB.Like.Delete(like); err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}
	ws.Respond(w, http.StatusNoContent, nil)
}
