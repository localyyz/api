package collection

import (
	"fmt"
	"net/http"
	"strings"

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

	productCond := db.Or(
		db.Raw("p.id IN (SELECT product_id FROM collection_products WHERE collection_id = ?)", collection.ID),
	)
	if collection.PlaceIDs != nil {
		placeIDs := make([]int64, len(*collection.PlaceIDs))
		for i, v := range *collection.PlaceIDs {
			placeIDs[i] = int64(v)
		}
		productCond = productCond.Or(db.Cond{"p.place_id IN": placeIDs})
	}
	if collection.Categories != nil {
		args := make([]string, len(*collection.Categories))
		for i, v := range *collection.Categories {
			args[i] = fmt.Sprintf(`'{"value":"%s"}'`, v)
		}
		//NOTE syntax: category @> any (ARRAY ['{"value":"bikini"}', '{"value":"swimwear"}']::jsonb[]);
		productCond = productCond.Or(db.Raw(fmt.Sprintf("category @> any (ARRAY [%s]::jsonb[])", strings.Join(args, ","))))
	}
	cond := db.And(
		productCond,
		db.Cond{
			"p.gender": collection.Gender,
			"p.status": data.ProductStatusApproved,
		},
	)
	query := data.DB.Select(db.Raw("distinct p.*")).
		From("products p").
		Where(cond).
		OrderBy("p.score DESC", "p.created_at DESC")

	query = filterSort.UpdateQueryBuilder(query)
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
