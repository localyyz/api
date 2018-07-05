package presenter

import (
	"context"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/go-chi/render"
	db "upper.io/db.v3"
)

type LightningCollection struct {
	*data.Collection
	Products []render.Renderer `json:"products"`
}

func (c *LightningCollection) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

/*
	Calculates the percentage complete -> counts the number of checkouts for the products in the collection
	Selects one product -> retrieves the first product from the collection which is in stock
*/
func NewLightningCollectionList(ctx context.Context, collections []*data.Collection) []render.Renderer {
	list := []render.Renderer{}
	for _, c := range collections {
		//append to the final list to return
		list = append(list, NewLightningCollection(ctx, c))
	}

	return list
}

func NewLightningCollection(ctx context.Context, collection *data.Collection) *LightningCollection {
	presented := &LightningCollection{
		Collection: collection,
	}
	if collection.Status == data.CollectionStatusActive {
		cps, err := data.DB.CollectionProduct.FindByCollectionID(collection.ID)
		if err != nil {
			return presented
		}
		var productIDs []int64
		for _, p := range cps {
			productIDs = append(productIDs, p.ProductID)
		}
		products, err := data.DB.Product.FindAll(db.Cond{
			"id":     productIDs,
			"status": data.ProductStatusApproved,
		})
		presented.Products = NewProductList(ctx, products)
	}
	return presented
}
