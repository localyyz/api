package product

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

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

func ListFeaturedProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cursor := ctx.Value("cursor").(*api.Page)

	cond := db.Cond{
		db.Raw("fp.product_id % 7"): db.Eq(time.Now().Weekday() + 1),
	}
	if gender, ok := ctx.Value("session.gender").(data.UserGender); ok {
		cond["gender"] = gender
	}

	var products []*data.Product
	query := data.DB.Select(db.Raw("p.*")).
		From("feature_products fp").
		LeftJoin("products p").
		On("p.id = fp.product_id").
		Where(cond).
		OrderBy("fp.ordering")
	paginate := cursor.UpdateQueryBuilder(query)
	if err := paginate.All(&products); err != nil {
		render.Respond(w, r, err)
		return
	}
	cursor.Update(products)

	presented := presenter.NewProductList(ctx, products)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}
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
			db.Cond{"p.gender": product.Gender, "p.id <>": product.ID},
			db.Raw(fmt.Sprintf("category @> '%s'", string(rawCategory))),
		)
		// find the products
		query = data.DB.Select(
			db.Raw("distinct p.*"),
			db.Raw(data.ProductWeightWithID),
		).
			From("products p").
			LeftJoin("places pl").On("pl.id = p.place_id").
			Where(cond).
			OrderBy("_rank desc")
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

func ListOnsaleProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cursor := ctx.Value("cursor").(*api.Page)

	// Only show on sale products from place weight >= 5
	featuredPlaces, _ := data.DB.Place.FindFeaturedMerchants()
	placeIDs := make([]int64, len(featuredPlaces))
	for i, p := range featuredPlaces {
		placeIDs[i] = p.ID
	}

	dayOfWeekPlusOne := int(time.Now().Weekday()) + 1
	query := data.DB.Select("*").
		From(db.Raw(`(
			SELECT product_id, row_number() over (partition by place_id, product_id % ?) as rank
			FROM product_variants
			WHERE place_id IN ?
			AND prev_price != 0
			AND prev_price > price
			GROUP BY place_id, product_id
		) x`, dayOfWeekPlusOne, placeIDs)).
		Where(db.Cond{
			"rank": dayOfWeekPlusOne,
		}).
		OrderBy(db.Raw("product_id % ?", dayOfWeekPlusOne))
	paginator := cursor.UpdateQueryBuilder(query)

	rows, err := paginator.QueryContext(ctx)
	if err != nil {
		render.Respond(w, r, err)
		return
	}
	defer rows.Close()

	var productIDs []int64
	for rows.Next() {
		var pId int64
		var rank interface{}
		if err := rows.Scan(&pId, &rank); err != nil {
			lg.Warnf("error scanning query: %+v", err)
			break
		}
		productIDs = append(productIDs, pId)
	}
	if err := rows.Err(); err != nil {
		render.Respond(w, r, err)
		return
	}

	cond := db.Cond{"id": productIDs}
	if gender, ok := ctx.Value("session.gender").(data.UserGender); ok {
		cond["gender"] = gender
	}
	result := data.DB.Product.Find(cond).
		OrderBy(
			data.MaintainOrder("id", productIDs),
		)
	var products []*data.Product
	if err := result.All(&products); err != nil {
		render.Respond(w, r, err)
		return
	}
	cursor.Update(products)

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
