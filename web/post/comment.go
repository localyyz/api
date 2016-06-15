package post

import (
	"net/http"
	"strconv"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
	"bitbucket.org/moodie-app/moodie-api/web/utils"

	"github.com/pressly/chi"

	"golang.org/x/net/context"
)

func CommentCtx(next chi.Handler) chi.Handler {
	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		commentID, err := strconv.ParseInt(chi.URLParam(ctx, "commentID"), 10, 64)
		if err != nil {
			ws.Respond(w, http.StatusBadRequest, utils.ErrBadID)
			return
		}

		comment, err := data.DB.Comment.FindByID(commentID)
		if err != nil {
			ws.Respond(w, http.StatusInternalServerError, err)
			return
		}
		ctx = context.WithValue(ctx, "comment", comment)
		next.ServeHTTPC(ctx, w, r)
	}
	return chi.HandlerFunc(handler)
}

func ListPostComment(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	post := ctx.Value("post").(*data.Post)
	comments, err := data.DB.Comment.FindByPostID(post.ID)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}
	ws.Respond(w, 200, comments)
}

func GetComment(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	comment := ctx.Value("comment").(*data.Comment)
	ws.Respond(w, 200, comment)
}

func AddComment(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	user := ctx.Value("session.user").(*data.User)
	post := ctx.Value("post").(*data.Post)

	var payload struct {
		*data.Comment

		// Ignore
		ID        interface{} `json:"id,omitempty"`
		PostID    interface{} `json:"postId,omitempty"`
		CreatedAt interface{} `json:"createdAt,omitempty"`
	}
	if err := ws.Bind(r.Body, &payload); err != nil {
		ws.Respond(w, http.StatusBadRequest, err)
		return
	}
	newComment := payload.Comment
	newComment.UserID = user.ID
	newComment.PostID = post.ID

	if err := data.DB.Comment.Save(newComment); err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	ws.Respond(w, http.StatusCreated, newComment)
}

func DeleteComment(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	user := ctx.Value("session.user").(*data.User)
	comment := ctx.Value("comment").(*data.Comment)
	if comment.UserID != user.ID {
		ws.Respond(w, http.StatusBadRequest, utils.ErrBadAction)
		return
	}

	if err := data.DB.Comment.Delete(comment); err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	ws.Respond(w, http.StatusNoContent, nil)
}
