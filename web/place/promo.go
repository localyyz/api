package place

import (
	"net/http"
	"time"

	"github.com/pressly/chi/render"

	set "gopkg.in/fatih/set.v0"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"upper.io/db.v3"
)

// List promotions at a given place grouped by product
func ListPromo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)
	cursor := api.NewPage(r)

	var promos []*data.Promo
	query := data.DB.Promo.Find(
		db.Cond{
			"place_id":    place.ID,
			"start_at <=": time.Now().UTC(),
			"end_at >":    time.Now().UTC(),
			"status":      data.PromoStatusActive,
		},
	).Select("product_id").Group("product_id").OrderBy("product_id")

	query = cursor.UpdateQueryUpper(query)
	if err := query.All(&promos); err != nil {
		render.Respond(w, r, err)
		return
	}

	productIDs := set.New()
	for _, p := range promos {
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
