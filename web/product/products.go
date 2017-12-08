package product

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/pressly/lg"

	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
)

func ProductCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		productID, err := strconv.ParseInt(chi.URLParam(r, "productID"), 10, 64)
		if err != nil {
			render.Render(w, r, api.ErrBadID)
			return
		}

		product, err := data.DB.Product.FindByID(productID)
		if err != nil {
			render.Respond(w, r, err)
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, "product", product)
		lg.SetEntryField(ctx, "product_id", product.ID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(handler)
}

func ListFeaturedProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	query := data.DB.Select("p.*").
		From("feature_products fp").
		LeftJoin("products p").
		On("fp.product_id = p.id").
		OrderBy("ordering")

	var products []*data.Product
	if err := query.All(&products); err != nil {
		render.Respond(w, r, err)
		return
	}

	w.Header().Add("X-Item-Total", fmt.Sprintf("%d", len(products)))
	presented := presenter.NewProductList(ctx, products)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}
}

func ListRecentProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// select the first row in each place_id group ordered by created_at
	q := `
		select *
		from (
			select row_number() over (partition by p.place_id order by p.created_at desc) as r, p.*
			from products p
			left join places pl on p.place_id = pl.id
			where p.created_at > now()::date - 1
			and pl.status = 3
		) x
		where x.r = 1
	`
	iter := data.DB.Iterator(q)
	defer iter.Close()
	var products []*data.Product
	if err := iter.All(&products); err != nil {
		render.Respond(w, r, err)
		return
	}

	w.Header().Add("X-Item-Total", fmt.Sprintf("%d", len(products)))
	presented := presenter.NewProductList(ctx, products)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}
}

func GetProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	product := ctx.Value("product").(*data.Product)
	render.Render(w, r, presenter.NewProduct(ctx, product))
}

func GetVariant(w http.ResponseWriter, r *http.Request) {
	product := r.Context().Value("product").(*data.Product)
	q := r.URL.Query()

	// look up variant by color and size
	var variant *data.ProductVariant
	err := data.DB.ProductVariant.Find(
		db.And(
			db.Cond{"product_id": product.ID},
			db.Raw("lower(etc->>'color') = ?", q.Get("color")),
			db.Raw("lower(etc->>'size') = ?", q.Get("size")),
		),
	).One(&variant)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Respond(w, r, variant)
}
