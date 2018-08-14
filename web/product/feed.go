package product

import (
	"net/http"

	"github.com/go-chi/render"

	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
)

func ListRandomProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cursor := ctx.Value("cursor").(*api.Page)
	filterSort := ctx.Value("filter.sort").(*api.FilterSort)

	cursor.Limit = 20 // hard coded
	query := data.DB.Select("p.*").
		From("products p").
		Where(db.Cond{
			"status": data.ProductStatusApproved,
			"score":  db.Gte(4),
		}).
		OrderBy(db.Raw("RANDOM()"))
	query = filterSort.UpdateQueryBuilder(query)

	var products []*data.Product
	if !filterSort.HasFilter() {
		paginate := cursor.UpdateQueryBuilder(query)
		if err := paginate.All(&products); err != nil {
			render.Respond(w, r, err)
			return
		}
		cursor.Update(products)
	}

	presented := presenter.NewProductList(ctx, products)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}
}
