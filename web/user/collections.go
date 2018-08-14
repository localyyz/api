package user

import (
	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"context"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/pkg/errors"
	"github.com/pressly/lg"
	"net/http"
	"strconv"
	"time"
	"upper.io/db.v3"
)

type newCollection struct {
	Title string `json:"title"`
}

func (c *newCollection) Bind(r *http.Request) error {
	return nil
}

type newCollectionProduct struct {
	ProductID int64 `json:"productId"`
}

func (c *newCollectionProduct) Bind(r *http.Request) error {
	return nil
}

// UserCollectionCtx finds the collection in the db and stores it in context
func UserCollectionCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		collectionID, err := strconv.ParseInt(chi.URLParam(r, "collectionID"), 10, 64)
		if err != nil {
			render.Render(w, r, api.ErrBadID)
			return
		}

		ctx := r.Context()
		user := ctx.Value("session.user").(*data.User)
		userCollection, err := data.DB.UserCollection.FindByID(user.ID, collectionID)
		if err != nil {
			render.Render(w, r, api.ErrBadID)
			return
		}
		ctx = context.WithValue(ctx, "user.collection", userCollection)
		lg.SetEntryField(ctx, "user_collection_id", userCollection.ID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

// CreateUserCollection creates a new user collection
func CreateUserCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)

	var payload newCollection
	if err := render.Bind(r, &payload); err != nil {
		render.Respond(w, r, api.ErrInvalidRequest(err))
		return
	}

	userCollection := data.UserCollection{
		UserID: user.ID,
		Title:  payload.Title,
	}

	err := data.DB.UserCollection.Create(userCollection)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// ListUserCollections returns a list of the user's collections
func ListUserCollections(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)

	userCollections, err := data.DB.UserCollection.FindByUserID(user.ID)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Respond(w, r, presenter.NewUserCollectionList(ctx, userCollections))
}

// GetUserCollection returns a specific user collection
func GetUserCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	collection := ctx.Value("user.collection").(*data.UserCollection)

	render.Render(w, r, presenter.NewUserCollection(ctx, collection))
}

// GetUserCollectionProducts returns the products from a specific collection
func GetUserCollectionProducts(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	collection := ctx.Value("user.collection").(*data.UserCollection)
	cursor := ctx.Value("cursor").(*api.Page)

	res := data.DB.UserCollectionProduct.Find(db.Cond{"collection_id": collection.ID, "deleted_at": db.IsNull()}).OrderBy("created_at DESC")
	paginate := cursor.UpdateQueryUpper(res)

	var userCollectionProducts []*data.UserCollectionProduct
	err := paginate.All(&userCollectionProducts)
	if err != nil {
		render.Respond(w, r, err)
		return
	}
	cursor.Update(userCollectionProducts)

	var productIDs []int64
	for _, uP := range userCollectionProducts {
		productIDs = append(productIDs, uP.ProductID)
	}

	products, err := data.DB.Product.FindAll(db.Cond{"id": productIDs})
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Respond(w, r, presenter.NewProductList(ctx, products))
}

// UpdateUserCollection updates the title of an existing collection
func UpdateUserCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	collection := ctx.Value("user.collection").(*data.UserCollection)

	var payload newCollection
	if err := render.Bind(r, &payload); err != nil {
		render.Respond(w, r, api.ErrInvalidRequest(err))
		return
	}

	if payload.Title == collection.Title {
		return
	}

	collection.Title = payload.Title
	err := data.DB.UserCollection.Save(collection)
	if err != nil {
		render.Respond(w, r, err)
		return
	}
}

// AddTouserCollectionProducts adds a product to an existing collection
func AddToUserCollectionProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	collection := ctx.Value("user.collection").(*data.UserCollection)

	var payload newCollectionProduct
	if err := render.Bind(r, &payload); err != nil {
		render.Respond(w, r, api.ErrInvalidRequest(err))
		return
	}

	if exists, _ := data.DB.Product.Find(db.Cond{"id": payload.ProductID}).Exists(); !exists {
		render.Respond(w, r, errors.New("Product does not exist"))
		return
	}

	var userCollectionProduct *data.UserCollectionProduct
	res := data.DB.UserCollectionProduct.Find(db.Cond{"collection_id": collection.ID, "product_id": payload.ProductID})
	if exists, _ := res.Exists(); exists {
		res.One(&userCollectionProduct)

		// the product exists and hasn't been deleted
		if userCollectionProduct.DeletedAt == nil {
			return
		}

		// the product exists and was deleted so "undelete" it
		_, err := data.DB.Update("user_collection_products").Set("deleted_at=NULL").Where(db.Cond{"collection_id": collection.ID, "product_id": payload.ProductID, "deleted_at": db.IsNotNull()}).Exec()
		if err != nil {
			render.Respond(w, r, err)
			return

		}
	} else {
		c := data.UserCollectionProduct{
			CollectionID: collection.ID,
			ProductID:    payload.ProductID,
		}

		err := data.DB.UserCollectionProduct.Create(c)
		if err != nil {
			render.Respond(w, r, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func DeleteUserCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	collection := ctx.Value("user.collection").(*data.UserCollection)

	// soft delete
	now := time.Now()
	collection.DeletedAt = &now

	err := data.DB.UserCollection.Save(collection)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	// delete everything from user_collection_products
	_, err = data.DB.Update("user_collection_products").Set("deleted_at=NOW()").Where(db.Cond{"collection_id": collection.ID, "deleted_at": db.IsNull()}).Exec()
	if err != nil {
		render.Respond(w, r, err)
		return
	}
}
