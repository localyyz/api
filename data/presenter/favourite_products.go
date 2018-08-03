package presenter

import (
	"bitbucket.org/moodie-app/moodie-api/data"
	"context"
	"github.com/go-chi/render"
	"upper.io/db.v3"
)

func FavouriteProductList(ctx context.Context, favProds []*data.FavouriteProduct) []render.Renderer {
	var productIDs []int64
	for _, c := range favProds {
		productIDs = append(productIDs, c.ProductID)
	}

	var products []*data.Product
	res := data.DB.Product.Find(db.Cond{"id": productIDs}).OrderBy(data.MaintainOrder("id", productIDs))
	res.All(&products)
	return NewProductList(ctx, products)
}
