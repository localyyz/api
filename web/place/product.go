package place

import (
	"net/http"

	"github.com/go-chi/render"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"upper.io/db.v3"
)

// List products at a given place
func ListProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)

	query := data.DB.
		Select(db.Raw("distinct p.*")).
		From("products p").
		LeftJoin("product_variants pv").
		On("pv.product_id = p.id").
		Where(
			db.Cond{
				"pv.limits >": 0,
				"p.place_id":  place.ID,
			},
		).
		GroupBy("p.id").
		OrderBy("-p.id")
	cursor := ctx.Value("cursor").(*api.Page)
	paginate := cursor.UpdateQueryBuilder(query)

	var products []*data.Product
	if err := paginate.All(&products); err != nil {
		render.Respond(w, r, err)
		return
	}
	cursor.Update(products)

	// count query
	row, _ := data.DB.
		Select(db.Raw("count(distinct pv.product_id)")).
		From("product_variants pv").
		Where(db.Cond{
			"pv.limits >": 0,
			"pv.place_id": place.ID,
		}).QueryRow()
	row.Scan(&cursor.ItemTotal)

	presented := presenter.NewProductList(ctx, products)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}
}
