package product

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/pressly/lg"

	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/data/stash"
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
				"p.place_id":                   product.PlaceID,
				"p.gender":                     product.Gender,
				"p.id":                         db.NotEq(product.ID),
				"p.status":                     data.ProductStatusApproved,
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

func ListTrending(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filterSort := ctx.Value("filter.sort").(*api.FilterSort)
	cursor := ctx.Value("cursor").(*api.Page)

	resp, err := http.DefaultClient.Get("http://reporter:5339/trend")
	if err != nil {
		render.Respond(w, r, []struct{}{})
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	var result presenter.ProductTrend
	if err := json.Unmarshal(b, &result); err != nil {
		render.Respond(w, r, err)
		return
	}

	query := data.DB.Select("p.*").
		From("products p").
		Where(db.Cond{
			"p.status": data.ProductStatusApproved,
			"p.id":     result.IDs,
		}).
		OrderBy(data.MaintainOrder("p.id", result.IDs))
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

func AddFavouriteProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user := ctx.Value("session.user").(*data.User)
	product := ctx.Value("product").(*data.Product)

	favProd := data.FavouriteProduct{ProductID: product.ID, UserID: user.ID}

	err := data.DB.FavouriteProduct.Create(favProd)
	if err != nil {
		render.Respond(w, r, err)
	}
}

func DeleteFavouriteProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user := ctx.Value("session.user").(*data.User)
	product := ctx.Value("product").(*data.Product)

	err := data.DB.FavouriteProduct.Find(db.Cond{"user_id": user.ID, "product_id": product.ID}).Delete()
	if err != nil {
		render.Respond(w, r, err)
	}
}

func DeleteFromAllCollections(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	product := ctx.Value("product").(*data.Product)
	user := ctx.Value("session.user").(*data.User)

	// delete from favourite product
	DeleteFavouriteProduct(w, r)

	var userColls []*data.UserCollection
	err := data.DB.Select("uc.*").
		From("user_collections as uc").
		LeftJoin("user_collection_products as ucp").
		On("ucp.collection_id = uc.id").
		Where(db.Cond{"uc.user_id": user.ID, "ucp.product_id": product.ID}).
		All(&userColls)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	// delete a product from all the user's collections
	res, err := data.DB.Exec("update user_collection_products as ucp set deleted_at = NOW()"+
		" from user_collections as uc"+
		" where ucp.collection_id = uc.id"+
		" and uc.user_id = $1"+
		" and ucp.product_id = $2",
		user.ID, product.ID)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	if affected, _ := res.RowsAffected(); affected > 0 {
		for _, coll := range userColls {
			stash.DecrUserCollProdCount(coll.ID)
			saving := product.Price * product.DiscountPct
			stash.DecrUserCollSavings(coll.ID, saving)
		}
	}
}

func DeleteProductFromCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	product := ctx.Value("product").(*data.Product)
	user := ctx.Value("session.user").(*data.User)
	collection := ctx.Value("user.collection").(*data.UserCollection)

	// delete a product from the user's specific collection
	res, err := data.DB.Exec("update user_collection_products as ucp set deleted_at = NOW()"+
		" from user_collections as uc"+
		" where ucp.collection_id = uc.id"+
		" and uc.user_id = $1"+
		" and ucp.product_id = $2"+
		" and ucp.collection_id = $3",
		user.ID, product.ID, collection.ID)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	if affected, _ := res.RowsAffected(); affected == 1 {
		// update the collection
		collection.UpdatedAt = data.GetTimeUTCPointer()
		err = data.DB.UserCollection.Save(collection)
		if err != nil {
			render.Respond(w, r, err)
			return
		}

		stash.DecrUserCollProdCount(collection.ID)

		savings := product.Price * product.DiscountPct
		stash.DecrUserCollSavings(collection.ID, savings)
	}
}
