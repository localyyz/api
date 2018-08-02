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

	products, _ := data.DB.Product.FindAll(db.Cond{"id": productIDs})

	return NewProductList(ctx, products)
}
