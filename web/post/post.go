package post

import (
	"context"
	"net/http"
	"strconv"

	"upper.io/db"

	"bitbucket.org/moodie-app/moodie-api/data"
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
	post := r.Context().Value("post").(*data.Post)

	place, err := data.DB.Place.FindByID(post.PlaceID)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}
	promo, err := data.DB.Promo.FindByID(post.PromoID)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	user, err := data.DB.User.FindByID(post.UserID)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}
	resp := data.PostPresenter{Post: post, Place: place, Promo: promo, User: user}

	ws.Respond(w, http.StatusOK, resp)
}

func CreatePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)

	var payload struct {
		data.Post

		// required...
		// PromoID int64 ?
		GooglePlaceID string `json:"googlePlaceId,required"`

		// Ignore
		ID          interface{} `json:"id,omitempty"`
		PlaceID     interface{} `jsoN:"placeId,omitempty"`
		UserID      interface{} `json:"userId,omitempty"`
		Comments    interface{} `json:"comments,omitempty"`
		Likes       interface{} `json:"comments,omitempty"`
		PromoStatus interface{} `json:"promoStatus,omitempty"`
		CreatedAt   interface{} `json:"createdAt,omitempty"`
		UpdatedAt   interface{} `json:"updatedAt,omitempty"`
	}
	if err := ws.Bind(r.Body, &payload); err != nil {
		ws.Respond(w, http.StatusBadRequest, err)
		return
	}

	newPost := &payload.Post
	newPost.UserID = user.ID

	var (
		place *data.Place
		err   error
	)
	place, err = data.DB.Place.FindByGoogleID(payload.GooglePlaceID)
	if err != nil {
		if err != db.ErrNoMoreRows {
			ws.Respond(w, http.StatusInternalServerError, err)
			return
		}
		place, err = data.GetPlaceDetail(ctx, payload.GooglePlaceID)
		if err != nil {
			ws.Respond(w, http.StatusInternalServerError, err)
			return
		}
		locale, err := data.GetLocale(ctx, &user.Geo)
		if err != nil {
			ws.Respond(w, http.StatusInternalServerError, err)
			return
		}
		place.LocaleID = locale.ID
		if err := data.DB.Place.Save(place); err != nil {
			ws.Respond(w, http.StatusInternalServerError, err)
			return
		}
	}

	// TODO: frontend work flow...
	// for now, let's just find the promo attached to the location
	//   picked by this post. auto apply promotion
	promo, err := data.DB.Promo.FindByPlaceID(place.ID)
	if err != nil && err != db.ErrNoMoreRows {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}
	if promo != nil {
		newPost.PromoID = promo.ID
	}

	newPost.PlaceID = place.ID
	if err := data.DB.Post.Save(newPost); err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	ws.Respond(w, http.StatusCreated, newPost)
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

	resp := struct {
		*data.User
		*data.Post
	}{User: user, Post: updatePost}

	ws.Respond(w, http.StatusCreated, resp)
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
