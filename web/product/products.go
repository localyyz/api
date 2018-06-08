package product

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/pressly/lg"

	db "upper.io/db.v3"
	"upper.io/db.v3/lib/sqlbuilder"

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

func ListRelatedProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	product := ctx.Value("product").(*data.Product)

	// if the product is featured, return "related" featured products
	exists, _ := data.DB.FeatureProduct.Find(db.Cond{"product_id": product.ID}).Exists()
	var query sqlbuilder.Selector
	if exists {
		cond := db.Cond{
			"product_id": db.NotEq(product.ID),
		}
		if gender, ok := ctx.Value("session.gender").(data.UserGender); ok {
			cond["gender"] = gender
		}
		query = data.DB.Select(db.Raw("p.*")).
			From("feature_products fp").
			LeftJoin("products p").
			On("p.id = fp.product_id").
			Where(cond).
			OrderBy("fp.ordering")
	} else {
		if product.Category.Value == "" {
			render.Respond(w, r, []struct{}{})
			return
		}
		rawCategory, _ := json.Marshal(product.Category)
		cond := db.And(
			db.Cond{
				"p.place_id": product.PlaceID,
				"p.gender":   product.Gender,
				"p.id <>":    product.ID,
			},
			db.Raw(fmt.Sprintf("category @> '%s'", string(rawCategory))),
		)
		// find the products
		query = data.DB.Select("p.*").
			From("products p").
			Where(cond).
			OrderBy("p.score desc", "p.created_at desc")
	}
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

func GetProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	product := ctx.Value("product").(*data.Product)
	render.Render(w, r, presenter.NewProduct(ctx, product))
}

func ListProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filterSort := ctx.Value("filter.sort").(*api.FilterSort)
	cursor := ctx.Value("cursor").(*api.Page)

	query := data.DB.Select("p.*").
		From("products p").
		Where(db.Cond{
			"status": data.ProductStatusApproved,
		}).
		OrderBy("-score")
	query = filterSort.UpdateQueryBuilder(query)
	paginate := cursor.UpdateQueryBuilder(query)

	var products []*data.Product
	if err := paginate.All(&products); err != nil {
		render.Respond(w, r, err)
		return
	}
	cursor.Update(products)

	render.RenderList(w, r, presenter.NewProductList(ctx, products))
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
