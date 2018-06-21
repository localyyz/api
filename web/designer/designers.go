package designer

import (
	"context"
	"net/http"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
)

var (
	featuredDesigners = []string{
		"calvin klein",
		"converse",
		"the north face",
		"puma",
		"timberland",
		"nike",
		"michael kors",
		"prada",
		"asics",
		"bottega veneta",
		"gucci",
		"valentino",
		"tissot",
		"coach",
		"levi's",
		"adidas",
		"tommy hilfiger",
		"louis vuitton",
		"dolce & gabbana",
		"diesel",
		"versace",
		"guess",
	}
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(api.FilterSortCtx)

	r.Get("/", List)
	r.Get("/featured", ListFeatured)
	r.Route("/{designer}", func(r chi.Router) {
		r.Use(DesignerCtx)
		r.Get("/products", ListProducts)
	})

	return r
}

func DesignerCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		designer := chi.URLParam(r, "designer")

		ctx = context.WithValue(ctx, "designer", strings.ToLower(designer))
		lg.SetEntryField(ctx, "designer", strings.ToLower(designer))
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cursor := ctx.Value("cursor").(*api.Page)

	cond := db.Cond{
		"p.deleted_at": nil,
		"p.brand":      db.NotEq(""),
		"p.status":     data.ProductStatusApproved,
	}
	var products []*data.Product
	query := data.DB.Select(db.Raw("distinct lower(p.brand) as brand")).
		From("products p").
		Where(cond).
		GroupBy(db.Raw("lower(p.brand)")).
		OrderBy(db.Raw("lower(p.brand)"))

	paginate := cursor.UpdateQueryBuilder(query)
	if err := paginate.All(&products); err != nil {
		render.Respond(w, r, err)
		return
	}
	cursor.Update(products)

	presented := presenter.NewDesignerList(ctx, products)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}
}

func ListFeatured(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	productBrands := make([]*data.Product, len(featuredDesigners))
	for i, d := range featuredDesigners {
		productBrands[i] = &data.Product{Brand: d}
	}
	render.RenderList(w, r, presenter.NewDesignerList(ctx, productBrands))
}

func ListProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	designer := ctx.Value("designer").(string)
	cursor := ctx.Value("cursor").(*api.Page)
	filterSort := ctx.Value("filter.sort").(*api.FilterSort)

	query := data.DB.Select("p.*").
		From("products p").
		Where(db.Cond{
			"p.deleted_at":           nil,
			"p.status":               data.ProductStatusApproved,
			db.Raw("lower(p.brand)"): strings.ToLower(designer),
		}).
		OrderBy("p.score DESC", "p.id DESC")

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
