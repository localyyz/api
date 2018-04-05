package product

import (
	"net/http"
	"strconv"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"github.com/go-chi/render"
	db "upper.io/db.v3"
)

func ListHistoryProduct(w http.ResponseWriter, r *http.Request) {
	rawProductIds := r.URL.Query().Get("productIds")
	if len(rawProductIds) == 0 {
		render.Respond(w, r, []struct{}{})
		return
	}

	// parse raw productIds, comma separated
	strProductIds := strings.Split(rawProductIds, ",")
	var productIDs []int64
	for _, p := range strProductIds {
		pID, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			continue
		}
		productIDs = append(productIDs, pID)
	}

	// bulk fetch products
	products, err := data.DB.Product.FindAll(db.Cond{"id": productIDs})
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Respond(w, r, presenter.NewProductList(r.Context(), products))
}
