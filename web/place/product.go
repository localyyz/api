package place

import (
	"net/http"

	"github.com/pressly/chi/render"

	set "gopkg.in/fatih/set.v0"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"upper.io/db.v3"
)

// List products at a given place
func ListProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)
	cursor := api.NewPage(r)

	var variants []*data.Promo
	query := data.DB.Promo.Find(
		db.Cond{
			"place_id": place.ID,
			"status":   data.PromoStatusActive,
		}).
		Select("product_id").
		Group("product_id").
		OrderBy("product_id")

	query = cursor.UpdateQueryUpper(query)
	if err := query.All(&variants); err != nil {
		render.Respond(w, r, err)
		return
	}

	productIDs := set.New()
	for _, p := range variants {
		productIDs.Add(p.ProductID)
	}
	if productIDs.Size() == 0 {
		render.Respond(w, r, []struct{}{})
		return
	}

	products, err := data.DB.Product.FindAll(db.Cond{"id": productIDs.List()})
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	presented := presenter.NewProductList(ctx, products)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}

}
