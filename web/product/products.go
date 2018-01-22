package product

import (
	"context"
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

func ProductGenderCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		genderRaw := chi.URLParam(r, "gender")

		gender := new(data.ProductGender)
		if err := gender.UnmarshalText([]byte(genderRaw)); err != nil {
			render.Respond(w, r, api.ErrInvalidRequest(err))
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, "product.gender", *gender)
		next.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(handler)
}

func ListGenderProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	gender := ctx.Value("product.gender").(data.ProductGender)
	cursor := ctx.Value("cursor").(*api.Page)

	var products []*data.Product
	cond := db.And(db.Cond{"gender": gender})
	if extraCond, ok := ctx.Value("product.filter").(db.Cond); ok && len(extraCond) > 0 {
		cond = cond.And(extraCond)
		query := data.DB.Select(db.Raw("distinct p.*")).
			From("products p").
			LeftJoin("product_tags pt").
			On("p.id = pt.product_id").
			Where(cond).
			OrderBy("-id")
		paginate := cursor.UpdateQueryBuilder(query)
		if err := paginate.All(&products); err != nil {
			render.Respond(w, r, err)
			return
		}
	} else {
		cond = cond.And(db.Raw(`not (category @> '{"type": "lingerie"}')`))
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

	// fetch the product's tags
	relatedTags, err := data.DB.ProductTag.FindAll(db.Cond{
		"product_id": product.ID,
		"type": []data.ProductTagType{
			data.ProductTagTypeCategory,
			data.ProductTagTypeBrand,
		},
	})
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	if len(relatedTags) == 0 {
		render.Respond(w, r, []struct{}{})
		return
	}

	relatedCond := db.Cond{
		"p.gender": product.Gender,
		"p.id <>":  product.ID,
	}
	if len(relatedTags) == 2 {
		// if both category and brand are found, use category
		for _, t := range relatedTags {
			if t.Type == data.ProductTagTypeCategory {
				// found a suitable category
				relatedCond["pt.type"] = t.Type
				relatedCond["pt.value"] = t.Value
				break
			}
		}
	} else {
		// beggers are not chosers
		relatedCond["pt.type"] = relatedTags[0].Type
		relatedCond["pt.value"] = relatedTags[0].Value
	}

	// find the products
	query := data.DB.Select(db.Raw("distinct p.*")).
		From("products p").
		LeftJoin("product_tags pt").
		On("pt.product_id = p.id").
		Where(relatedCond).
		OrderBy("p.id desc")
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
			select row_number() over (partition by p.place_id order by p.created_at desc) as r, p.*
			from products p
			left join places pl on p.place_id = pl.id
			where p.created_at > now()::date - 7
			and pl.status = 3
			and not (category @> '{"type": "lingerie"}')
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
