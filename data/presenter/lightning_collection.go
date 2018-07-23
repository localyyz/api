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
	Products []*Product `json:"products"`
}

func (c *LightningCollection) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func NewDealList(ctx context.Context, collections []*data.Collection) []render.Renderer {
	list := []render.Renderer{}
	for _, c := range collections {
		//append to the final list to return
		list = append(list, NewDeal(ctx, c))
	}
	return list
}

func NewDeal(ctx context.Context, collection *data.Collection) *LightningCollection {
	presented := &LightningCollection{
		Collection: collection,
	}
	if collection.Status == data.CollectionStatusActive || collection.Status == data.CollectionStatusInactive {
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
		presented.Products = newProductList(ctx, products)
	}
	return presented
}
