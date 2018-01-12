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

	var featured []*data.FeatureProduct
	if err := data.DB.FeatureProduct.Find().
		OrderBy("ordering").All(&featured); err != nil {
		render.Respond(w, r, err)
		return
	}

	var productIDs []int64
	for _, p := range featured {
		productIDs = append(productIDs, p.ProductID)
	}

	products, err := data.DB.Product.FindAll(db.Cond{"id": productIDs})
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	w.Header().Add("X-Item-Total", fmt.Sprintf("%d", len(products)))
	presented := presenter.NewProductList(ctx, products)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}
}

func ListRelatedProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	product := ctx.Value("product").(*data.Product)

	// fetch the product's category
	category, err := data.DB.ProductTag.FindOne(db.Cond{
		"product_id": product.ID,
		"type":       data.ProductTagTypeCategory,
	})
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	// find the products
	query := data.DB.Select(db.Raw("distinct p.*")).
		From("products p").
		LeftJoin("product_tags pt").
		On("pt.product_id = p.id").
		Where(db.Cond{
			"pt.type":  data.ProductTagTypeCategory,
			"pt.value": category.Value,
			"p.gender": product.Gender,
			"p.id <>":  product.ID,
		})
	cursor := ctx.Value("cursor").(*api.Page)
	paginate := cursor.UpdateQueryBuilder(query)
	var relatedProducts []*data.Product
	if err := paginate.All(&relatedProducts); err != nil {
		render.Respond(w, r, err)
		return
	}

	cursor.Update(relatedProducts)
	presented := presenter.NewProductList(ctx, relatedProducts)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}
}

func ListRecentProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cursor := ctx.Value("cursor").(*api.Page)

	// select the first row in each place_id group ordered by created_at
	// TODO: is this bad? probably
	q := fmt.Sprintf(`select *
		from (
			select row_number() over (partition by p.place_id order by p.created_at desc) as r, p.*
			from products p
			left join places pl on p.place_id = pl.id
			where p.created_at > now()::date - 7
			and pl.status = 3
		) x
		where x.r = %d
		order by created_at desc
		limit 10`, cursor.Page)
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
