package post

import (
	"net/http"
	"strconv"

	"upper.io/db"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
	"bitbucket.org/moodie-app/moodie-api/web/utils"

	"github.com/goware/lg"
	"github.com/pressly/chi"
	"golang.org/x/net/context"
)

func PostCtx(next chi.Handler) chi.Handler {
	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		postID, err := strconv.ParseInt(chi.URLParam(ctx, "postID"), 10, 64)
		if err != nil {
			ws.Respond(w, http.StatusBadRequest, utils.ErrBadID)
			return
		}

		post, err := data.DB.Post.FindByID(postID)
		if err != nil {
			ws.Respond(w, http.StatusInternalServerError, err)
			return
		}
		ctx = context.WithValue(ctx, "post", post)
		next.ServeHTTPC(ctx, w, r)
	}
	return chi.HandlerFunc(handler)
}

func GetPost(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	post := ctx.Value("post").(*data.Post)
	ws.Respond(w, http.StatusOK, post)
}

func CreatePost(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	user := ctx.Value("session.user").(*data.User)

	var payload struct {
		data.Post

		GooglePlaceID     string `json:"googlePlaceId,required"`
		GooglePlaceDetail string `json:"googlePlaceDetail,required"`

		// Ignore
		ID        interface{} `json:"id,omitempty"`
		UserID    interface{} `json:"userId,omitempty"`
		Comments  interface{} `json:"comments,omitempty"`
		Likes     interface{} `json:"comments,omitempty"`
		CreatedAt interface{} `json:"createdAt,omitempty"`
		UpdatedAt interface{} `json:"updatedAt,omitempty"`
	}
	if err := ws.Bind(r.Body, &payload); err != nil {
		ws.Respond(w, http.StatusBadRequest, err)
		return
	}

	newPost := &payload.Post
	newPost.UserID = user.ID

	if payload.GooglePlaceID != "" {
		var place *data.Place
		err := data.DB.Place.Find(db.Cond{"google_id": payload.GooglePlaceID}).One(&place)
		if err != nil {
			if err != db.ErrNoMoreRows {
				ws.Respond(w, http.StatusInternalServerError, err)
				return
			}
			place = &data.Place{
				GoogleID: payload.GooglePlaceID,
				Name:     payload.GooglePlaceDetail,
				LocaleID: 1, // for now.. TODO: need to grab the detail from google?
			}
			if err := data.DB.Place.Save(place); err != nil {
				ws.Respond(w, http.StatusInternalServerError, err)
				lg.Warn(err)
				return
			}
		}
		if place != nil && place.ID != 0 {
			newPost.PlaceID = place.ID
		}
	}

	if err := data.DB.Post.Save(newPost); err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	resp := struct {
		*data.User
		*data.Post
	}{User: user, Post: newPost}

	ws.Respond(w, http.StatusCreated, resp)
}

func UpdatePost(ctx context.Context, w http.ResponseWriter, r *http.Request) {
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

	resp := struct {
		*data.User
		*data.Post
	}{User: user, Post: updatePost}

	ws.Respond(w, http.StatusCreated, resp)
}

func DeletePost(ctx context.Context, w http.ResponseWriter, r *http.Request) {
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
