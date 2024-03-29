package category

import (
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	db "upper.io/db.v3"
)

func discountCtx(min, max float64) func(next http.Handler) http.Handler {
	return middleware.WithValue(
		"discountCond",
		db.Cond{
			"p.discount_pct": db.Between(min, max),
		},
	)
}

func ListDiscountProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cursor := ctx.Value("cursor").(*api.Page)
	filterSort := ctx.Value("filter.sort").(*api.FilterSort)

	cond := db.And(
		db.Cond{
			"p.status":     data.ProductStatusApproved,
			"p.deleted_at": nil,
			"m.pricing":    []string{"medium", "high"},
		},
	)
	if discountCond, ok := ctx.Value("discountCond").(db.Cond); ok {
		cond = cond.And(discountCond)
	}

	query := data.DB.Select("p.*").
		From("products p").
		LeftJoin("place_meta m").On("m.place_id = p.place_id").
		Where(cond).
		OrderBy("p.id desc")
	query = filterSort.UpdateQueryBuilder(query)

	if filterSort.HasFilter() {
		w.Write([]byte{})
		return
	}

	var products []*data.Product
	paginate := cursor.UpdateQueryBuilder(query)
	if err := paginate.All(&products); err != nil {
		render.Respond(w, r, err)
		return
	}
	cursor.Update(products)

	render.RenderList(w, r, presenter.NewProductList(ctx, products))
}
