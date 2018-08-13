package product

import (
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/render"
	db "upper.io/db.v3"
)

func ListFavourite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)
	cursor := ctx.Value("cursor").(*api.Page)

	res := data.DB.FavouriteProduct.Find(
		db.Cond{
			"user_id": user.ID,
		},
	).OrderBy("created_at DESC")
	paginate := cursor.UpdateQueryUpper(res)

	var favProducts []*data.FavouriteProduct
	err := paginate.All(&favProducts)
	if err != nil && err != db.ErrNoMoreRows {
		render.Respond(w, r, err)
		return
	}
	cursor.Update(favProducts)

	presented := presenter.FavouriteProductList(ctx, favProducts)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}
}
