package product

import (
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/events"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/render"
	db "upper.io/db.v3"
)

func AddFavouriteProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)
	product := ctx.Value("product").(*data.Product)

	f := data.FavouriteProduct{
		ProductID: product.ID,
		UserID:    user.ID,
	}
	if err := data.DB.FavouriteProduct.Create(f); err != nil {
		render.Respond(w, r, err)
		return
	}

	// emit event
	connect.NATS.Emit(
		events.EvProductFavourited,
		presenter.ProductEvent{
			Product:  product,
			ViewerID: user.ID,
		},
	)
}

func DeleteFavouriteProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user := ctx.Value("session.user").(*data.User)
	product := ctx.Value("product").(*data.Product)

	err := data.DB.FavouriteProduct.Find(db.Cond{"user_id": user.ID, "product_id": product.ID}).Delete()
	if err != nil {
		render.Respond(w, r, err)
	}
}

func ListFavourite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)
	cursor := ctx.Value("cursor").(*api.Page)
	filterSort := ctx.Value("filter.sort").(*api.FilterSort)

	fpQuery := data.DB.FavouriteProduct.Find(
		db.Cond{
			"user_id": user.ID,
		},
	).OrderBy("created_at DESC")
	fpQuery = cursor.UpdateQueryUpper(fpQuery)

	var favProducts []*data.FavouriteProduct
	if err := fpQuery.All(&favProducts); err != nil {
		render.Respond(w, r, err)
		return
	}
	cursor.Update(favProducts)

	productIDs := make([]int64, len(favProducts))
	for i, fp := range favProducts {
		productIDs[i] = fp.ProductID
	}

	query := data.DB.Select("p.*").
		From("products p").
		Where(db.Cond{
			"id":     productIDs,
			"status": data.ProductStatusApproved,
		})
	query = filterSort.UpdateQueryBuilder(query)

	var products []*data.Product
	if err := query.All(&products); err != nil {
		render.Respond(w, r, err)
		return
	}

	presented := presenter.NewProductList(ctx, products)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}
}
