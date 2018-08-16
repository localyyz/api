package product

import (
	"fmt"
	"net/http"
	"time"

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

	hardCond := db.Raw(`p.tsv @@ (to_tsquery($$puma$$) ||
				to_tsquery('simple', $$puma:*$$) ||
				to_tsquery($$puma:*$$) ||
				to_tsquery('simple', $$puma$$) ||
				
				to_tsquery($$nike$$) ||
				to_tsquery('simple', $$nike:*$$) ||
				to_tsquery($$nike:*$$) ||
				to_tsquery('simple', $$nike$$) ||

				to_tsquery($$yeezy$$) ||
				to_tsquery('simple', $$yeezy:*$$) ||
				to_tsquery($$yeezy:*$$) ||
				to_tsquery('simple', $$yeezy$$) ||
				
				to_tsquery($$supreme$$) ||
				to_tsquery('simple', $$supreme:*$$) ||
				to_tsquery($$supreme:*$$) ||
				to_tsquery('simple', $$supreme$$) ||

				to_tsquery($$moschino$$) ||
				
				to_tsquery($$jordans$$) ||
				to_tsquery('simple', $$jordans:*$$) ||
				to_tsquery($$jordans:*$$) ||
				to_tsquery('simple', $$jordans$$))
	`)
	cond := db.And(
		db.Cond{
			"p.status": data.ProductStatusApproved,
			"p.score":  db.Gte(4),
			db.Raw("p.category->>'type'"): []data.CategoryType{
				data.CategoryShoe,
				data.CategorySneaker,
				data.CategoryApparel,
			},
		},
		hardCond,
	)

	t := time.Now().Truncate(time.Hour).Unix()
	cursor.Limit = 20 // hard coded
	cursor.ItemTotal = 10000
	query := data.DB.Select("p.*").
		From("products p").
		Where(cond).
		OrderBy(db.Raw(fmt.Sprintf("%d %% id", t)), "-score")
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
