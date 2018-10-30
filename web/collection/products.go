package collection

import (
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/render"
	db "upper.io/db.v3"
)

func ListProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	collection := ctx.Value("collection").(*data.Collection)
	cursor := ctx.Value("cursor").(*api.Page)
	filterSort := ctx.Value("filter.sort").(*api.FilterSort)

	// default ->
	productCond := db.Cond{
		"p.id": db.Raw("(SELECT product_id FROM collection_products WHERE collection_id = ?)", collection.ID),
	}
	if collection.PlaceIDs != nil {
		placeIDs := make([]int64, len(*collection.PlaceIDs))
		for i, v := range *collection.PlaceIDs {
			placeIDs[i] = int64(v)
		}
		productCond = db.Cond{"p.place_id": placeIDs}
	}
	cond := db.And(
		productCond,
		db.Cond{"p.status": data.ProductStatusApproved},
	)
	query := data.DB.Select(db.Raw("distinct p.*")).
		From("products p").
		Where(cond).
		OrderBy("p.score DESC", "p.created_at DESC")
	query = filterSort.UpdateQueryBuilder(query)

	if filterSort.HasFilter() {
		w.Write([]byte{})
		return
	}

	paginate := cursor.UpdateQueryBuilder(query)
	var products []*data.Product
	if err := paginate.All(&products); err != nil {
		render.Respond(w, r, err)
		return
	}
	cursor.Update(products)

	presented := presenter.NewProductList(ctx, products)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}
}
