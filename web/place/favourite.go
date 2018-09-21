package place

import (
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/render"
	db "upper.io/db.v3"
)

func AddFavourite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)
	place := ctx.Value("place").(*data.Place)

	fp := data.FavouritePlace{
		PlaceID: place.ID,
		UserID:  user.ID,
	}
	if err := data.DB.FavouritePlace.Create(fp); err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Respond(w, r, fp)
}

func DeleteFavourite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)
	place := ctx.Value("place").(*data.Place)

	err := data.DB.FavouritePlace.Find(db.Cond{"user_id": user.ID, "place_id": place.ID}).Delete()
	if err != nil {
		render.Respond(w, r, err)
	}

	render.Status(r, http.StatusNoContent)
	if err != nil {
		return
	}
	render.Respond(w, r, "")
}

func ListFavourite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)
	cursor := ctx.Value("cursor").(*api.Page)

	favs, err := data.DB.FavouritePlace.FindByUserID(user.ID)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	placeIDs := make([]int64, len(favs))
	for i, f := range favs {
		placeIDs[i] = f.PlaceID
	}

	query := data.DB.Product.Find(db.Cond{"id": placeIDs}).OrderBy(data.MaintainOrder("id", placeIDs))
	query = cursor.UpdateQueryUpper(query)

	var places []*data.Place
	if err := query.All(&places); err != nil {
		render.Respond(w, r, err)
		return
	}
	cursor.Update(places)

	render.Respond(w, r, places)
}
