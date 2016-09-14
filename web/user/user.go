package user

import (
	"context"
	"net/http"
	"strconv"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
	"bitbucket.org/moodie-app/moodie-api/web/utils"

	"github.com/pressly/chi"

	"upper.io/db.v2"
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
	var (
		postID  []int64
		placeID []int64
		promoID []int64
	)
	for _, p := range points {
		if p.PostID != nil {
			postID = append(postID, *p.PostID)
		}
		placeID = append(placeID, p.PlaceID)
		promoID = append(promoID, p.PromoID)
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

	places, err := data.DB.Place.FindAll(db.Cond{"id": placeID})
	if err != nil {
		ws.Respond(w, http.StatusServiceUnavailable, err)
		return
	}
	placesMap := map[int64]*data.Place{}
	for _, p := range places {
		placesMap[p.ID] = p
	}

	promos, err := data.DB.Promo.FindAll(db.Cond{"id": promoID})
	if err != nil {
		ws.Respond(w, http.StatusServiceUnavailable, err)
		return
	}
	promosMap := map[int64]*data.Promo{}
	for _, p := range promos {
		promosMap[p.ID] = p
	}

	resp := make([]*data.UserPointPresenter, len(points))
	for i, p := range points {
		resp[i] = &data.UserPointPresenter{
			UserPoint: p,
		}

		promo, promoFound := promosMap[p.PromoID]
		if promoFound {
			resp[i].Promo = promo
		}

		if p.PostID != nil {
			resp[i].Post = postsMap[*p.PostID]
		}
		resp[i].Place = placesMap[p.PlaceID]
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
}
