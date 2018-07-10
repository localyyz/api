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
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/events"
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
	cursor := ctx.Value("cursor").(*api.Page)

	if product.Category.Value == "" {
		render.Respond(w, r, []struct{}{})
		return
	}

	// find the products
	query := data.DB.Select("p.*").
		From("products p").
		Where(
			db.Cond{
				"p.place_id": product.PlaceID,
				"p.gender":   product.Gender,
				"p.id":       db.NotEq(product.ID),
				db.Raw("p.category->>'value'"): product.Category.Value,
			}).
		OrderBy("p.score desc")
	paginate := cursor.UpdateQueryBuilder(query)

	var products []*data.Product
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

func GetProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	product := ctx.Value("product").(*data.Product)

	{
		evt := presenter.ProductEvent{Product: product}
		if sessionUser, ok := ctx.Value("session.user").(*data.User); ok {
			evt.ViewerID = sessionUser.ID
		}
		connect.NATS.Emit(events.EvProductViewed, evt)
	}

	render.Render(w, r, presenter.NewProduct(ctx, product))
}

func ListProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filterSort := ctx.Value("filter.sort").(*api.FilterSort)
	cursor := ctx.Value("cursor").(*api.Page)

	query := data.DB.Select("p.*").
		From("products p").
		Where(db.Cond{"p.status": data.ProductStatusApproved}).
		OrderBy("p.score DESC")
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
