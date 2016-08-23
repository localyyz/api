package post

import (
	"context"
	"net/http"
	"strconv"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
	"bitbucket.org/moodie-app/moodie-api/web/utils"

	"github.com/pressly/chi"
)

func PostCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		postID, err := strconv.ParseInt(chi.URLParam(r, "postID"), 10, 64)
		if err != nil {
			ws.Respond(w, http.StatusBadRequest, utils.ErrBadID)
			return
		}

		post, err := data.DB.Post.FindByID(postID)
		if err != nil {
			ws.Respond(w, http.StatusInternalServerError, err)
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, "post", post)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func GetPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	post := ctx.Value("post").(*data.Post)

	presented, err := presenter.NewPost(ctx, post)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	ws.Respond(w, http.StatusOK, presented)
}

func UpdatePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)
	post := ctx.Value("post").(*data.Post)
	if post.UserID != user.ID {
		ws.Respond(w, http.StatusBadRequest, utils.ErrBadAction)
		return
	}

	payload := struct {
		*data.Post

		// Ignore
		ID        interface{} `json:"id,omitempty"`
		UserID    interface{} `json:"userId,omitempty"`
		Comments  interface{} `json:"comments,omitempty"`
		Likes     interface{} `json:"comments,omitempty"`
		Filter    interface{} `json:"filter,omitempty"`
		ImageURL  interface{} `json:"imageUrl,omitempty"`
		CreatedAt interface{} `json:"createdAt,omitempty"`
		UpdatedAt interface{} `json:"updatedAt,omitempty"`
	}{Post: post}
	if err := ws.Bind(r.Body, &payload); err != nil {
		ws.Respond(w, http.StatusBadRequest, err)
		return
	}

	updatePost := payload.Post
	if err := data.DB.Post.Save(updatePost); err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	presented, err := presenter.NewPost(ctx, updatePost)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	ws.Respond(w, http.StatusCreated, presented)
}

func DeletePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	post := ctx.Value("post").(*data.Post)
	user := ctx.Value("session.user").(*data.User)
	if post.UserID != user.ID {
		ws.Respond(w, http.StatusBadRequest, utils.ErrBadAction)
		return
	}

	if err := data.DB.Post.Delete(post); err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	ws.Respond(w, http.StatusNoContent, nil)
}
