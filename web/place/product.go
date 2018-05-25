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
	cursor := ctx.Value("cursor").(*api.Page)
	place := ctx.Value("place").(*data.Place)

	query := data.DB.Product.
		Find(
			db.Cond{
				"place_id":   place.ID,
				"deleted_at": nil,
				"status":     data.ProductStatusApproved,
			},
		).OrderBy("score DESC", "created_at DESC")
	query = cursor.UpdateQueryUpper(query)

	var products []*data.Product
	if err := query.All(&products); err != nil {
		render.Respond(w, r, err)
		return
	}
	cursor.Update(products)

	presented := presenter.NewProductList(ctx, products)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}
}
