package presenter

import (
	"bitbucket.org/moodie-app/moodie-api/data"
	"context"
	"github.com/go-chi/render"
	"math"
	"net/http"
	"upper.io/db.v3"
)

type CollectionUser struct {
	*data.UserCollection
	TotalProducts uint64     `json:"totalProducts"`
	Products      []*Product `json:"products"`
	Savings       float64    `json:"savings"`
	Owner         string     `json:"owner"`

	ctx context.Context
}

func (c *CollectionUser) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func NewUserCollection(ctx context.Context, collection *data.UserCollection) *CollectionUser {
	c := &CollectionUser{
		UserCollection: collection,
		ctx:            ctx,
	}

	res := data.DB.UserCollectionProduct.Find(db.Cond{"collection_id": collection.ID, "deleted_at": db.IsNull()}).OrderBy("created_at DESC")

	// count the products in the collection
	totalProducts, _ := res.Count()
	c.TotalProducts = totalProducts

	// get the top 3 products
	var userCollectionProducts []*data.UserCollectionProduct
	res.Limit(3).All(&userCollectionProducts)

	var userCollectionProductIDs []int64
	for _, userCollectionProduct := range userCollectionProducts {
		userCollectionProductIDs = append(userCollectionProductIDs, userCollectionProduct.ProductID)
	}

	products, _ := data.DB.Product.FindAll(db.Cond{"id": userCollectionProductIDs})
	c.Products = newProductList(ctx, products)

	// count the total savings
	var totalSavings float64
	for _, product := range products {
		if product.DiscountPct > 0 {
			totalSavings += product.Price * product.DiscountPct
		}
	}
	// round to 2 decimal places
	c.Savings = math.Round(totalSavings*100) / 100

	// append the user name
	user := ctx.Value("session.user").(*data.User)
	c.Owner = user.Name

	return c
}

func NewUserCollectionList(ctx context.Context, collections []*data.UserCollection) []render.Renderer {
	list := []render.Renderer{}
	for _, collection := range collections {
		c := NewUserCollection(ctx, collection)
		list = append(list, c)
	}
	return list
}
