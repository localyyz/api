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

func ListGenderProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	gender := ctx.Value("product.gender").(data.ProductGender)
	cursor := ctx.Value("cursor").(*api.Page)

	var products []*data.Product
	cond := db.And(
		db.Cond{
			"gender":     gender,
			"deleted_at": nil,
		},
	)
	if extraCond, ok := ctx.Value("product.filter").(db.Cond); ok && len(extraCond) > 0 {
		cond = cond.And(extraCond)
		query := data.DB.Select(db.Raw("distinct *")).
			From("products").
			Where(cond).
			OrderBy("-id")
		paginate := cursor.UpdateQueryBuilder(query)
		if err := paginate.All(&products); err != nil {
			render.Respond(w, r, err)
			return
		}
	} else {
		cond = cond.And(
			db.Raw(`not (category @> '{"type": "lingerie"}')`),
			db.Raw(`not (category @> '{"type": "swimwear"}')`),
			db.Raw(`category ?? 'type'`),
		)
		query := data.DB.Product.Find(cond).OrderBy("-id")
		query = cursor.UpdateQueryUpper(query)
		if err := query.All(&products); err != nil {
			render.Respond(w, r, err)
			return
		}
	}
	cursor.Update(products)

	presented := presenter.NewProductList(ctx, products)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}
}

func ListFeaturedProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cursor := ctx.Value("cursor").(*api.Page)

	var products []*data.Product
	query := data.DB.Select(db.Raw("p.*")).
		From("feature_products fp").
		LeftJoin("products p").
		On("p.id = fp.product_id").
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
		query = data.DB.Select(db.Raw("p.*")).
			From("feature_products fp").
			LeftJoin("products p").
			On("p.id = fp.product_id").
			Where(db.Cond{"product_id": db.NotEq(product.ID)}).
			OrderBy("fp.ordering")
	} else {
		rawCategory, _ := json.Marshal(product.Category)
		relatedCond := db.And(
			db.Cond{"p.gender": product.Gender, "p.id <>": product.ID},
			db.Raw(fmt.Sprintf("category @> '%s'", string(rawCategory))),
		)
		// find the products
		query = data.DB.Select(db.Raw("distinct p.*")).
			From("products p").
			Where(relatedCond).
			OrderBy("p.id desc")
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

func ListRecentProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cursor := ctx.Value("cursor").(*api.Page)

	// select the first row in each place_id group ordered by created_at
	// TODO: is this bad? probably
	query := data.DB.Select("*").
		From(db.Raw(`(
			select row_number() over (partition by p.place_id order by p.weight desc, p.id desc) as r, p.*
			from products p
			left join places pl on p.place_id = pl.id
			where p.created_at > now()::date - 7
			and pl.status = 3
			and category ->> 'type' = 'apparel'
			and p.image_url != ''
			and p.deleted_at is null
		) x`)).
		Where("x.r = ?", cursor.Page).
		OrderBy("created_at desc").
		Limit(cursor.Limit)

	var products []*data.Product
	if err := query.All(&products); err != nil {
		render.Respond(w, r, err)
		return
	}
	cursor.Update(products)
	presented := presenter.NewProductList(ctx, products)
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

	query := data.DB.Select(db.Raw("product_id")).
		From("product_variants pv").
		Where(db.Cond{
			"place_id":      placeIDs,
			"prev_price !=": 0,
			"prev_price >":  db.Raw("price"),
		}).
		GroupBy("place_id", "product_id").
		OrderBy(data.MaintainOrder("place_id", placeIDs))
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
		if err := rows.Scan(&pId); err != nil {
			lg.Warnf("error scanning query: %+v", err)
			break
		}
		productIDs = append(productIDs, pId)
	}
	if err := rows.Err(); err != nil {
		render.Respond(w, r, err)
		return
	}

	result := data.DB.Product.Find(
		db.Cond{"id": productIDs},
	).OrderBy(
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
