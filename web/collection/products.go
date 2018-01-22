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

func ListProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	collection := ctx.Value("collection").(*data.Collection)
	cursor := ctx.Value("cursor").(*api.Page)

	productCond := db.Or(
		db.Raw("p.id IN (SELECT product_id FROM collection_products WHERE collection_id = ?)", collection.ID),
	)
	if collection.PlaceIDs != nil {
		productCond = productCond.Or(db.Cond{"p.place_id IN": collection.PlaceIDs})
	}
	if collection.Categories != nil {
		args := make([]string, len(*collection.Categories))
		for i, v := range *collection.Categories {
			args[i] = fmt.Sprintf(`'{"value":"%s"}'`, v)
		}
		//NOTE syntax: category @> any (ARRAY ['{"value":"bikini"}', '{"value":"swimwear"}']::jsonb[]);
		productCond = productCond.Or(db.Raw(fmt.Sprintf("category @> any (ARRAY [%s]::jsonb[])", strings.Join(args, ","))))
	}
	cond := db.And(productCond, db.Cond{"p.gender": collection.Gender})

	query := data.DB.Select(db.Raw("distinct p.*")).
		From("products p").
		Where(cond).
		OrderBy("p.id DESC")

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
