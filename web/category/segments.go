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

type segmentType string

const (
	segmentTypeLuxury   = "lux"
	segmentTypeBoutique = "bou"
	segmentTypeSmart    = "smt"
)

func segmentCtx(v segmentType) func(next http.Handler) http.Handler {
	var cond db.Cond
	switch v {
	case segmentTypeBoutique:
		cond = db.Cond{
			"m.collection": data.MerchantApprovalCollectionBoutique,
			"p.price":      db.Between(30, 430),
		}
	case segmentTypeLuxury:
		cond = db.Cond{
			"m.collection": data.MerchantApprovalCollectionLuxury,
			"p.price":      db.Gt(150),
		}
	case segmentTypeSmart:
		cond = db.Cond{
			"m.collection": data.MerchantApprovalCollectionSmart,
			"p.price":      db.Lt(150),
		}
	}
	return middleware.WithValue("segmentCond", cond)
}

func ListSegmentProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cursor := ctx.Value("cursor").(*api.Page)
	filterSort := ctx.Value("filter.sort").(*api.FilterSort)

	cond := db.And(
		db.Cond{
			"p.status":     data.ProductStatusApproved,
			"p.deleted_at": nil,
		},
	)
	if segmentCond, ok := ctx.Value("segmentCond").(db.Cond); ok {
		cond = cond.And(segmentCond)
	}

	query := data.DB.Select("p.*").
		From("products p").
		LeftJoin("merchant_approvals m").On("m.place_id = p.place_id").
		Where(cond).
		OrderBy("p.id desc", "p.score DESC")
	query = filterSort.UpdateQueryBuilder(query)

	var products []*data.Product
	paginate := cursor.UpdateQueryBuilder(query)
	if err := paginate.All(&products); err != nil {
		render.Respond(w, r, err)
		return
	}
	cursor.Update(products)

	render.RenderList(w, r, presenter.NewProductList(ctx, products))
}
