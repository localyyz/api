package place

import (
	"net/http"
	"time"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
	"upper.io/db.v3"
)

// List promotions at a given place grouped by product
func ListPromo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)
	cursor := ws.NewPage(r)

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
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	productIDs := make([]int64, len(promos))
	for i, p := range promos {
		productIDs[i] = p.ProductID
	}

	products, err := data.DB.Product.FindAll(db.Cond{"id": productIDs})
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	res := make([]*presenter.Product, len(products))
	for i, p := range products {
		res[i] = presenter.NewProduct(ctx, p).WithPromo()
	}

	ws.Respond(w, http.StatusOK, res, cursor.Update(products))
}
