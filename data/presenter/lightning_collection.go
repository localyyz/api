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

		presented := &LightningCollection{
			Collection: c,
		}
		if c.Status == data.CollectionStatusActive {
			cps, err := data.DB.CollectionProduct.FindByCollectionID(c.ID)
			if err != nil {
				return list
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

		//append to the final list to return
		list = append(list, presented)
	}

	return list
}

func NewLightningCollection(ctx context.Context, collection *data.Collection) render.Renderer {
	lightningCollection := &LightningCollection{}

	//setting the collection
	lightningCollection.Collection = collection

	var collectionProducts []*data.CollectionProduct
	err := data.DB.CollectionProduct.Find(db.Cond{"collection_id": collection.ID}).All(&collectionProducts)
	if err != nil {
		return nil
	}

	//iterate over the products from the collection
	var productIDs []int64
	for _, collectionProduct := range collectionProducts {
		productIDs = append(productIDs, collectionProduct.ProductID)
	}
	products, err := data.DB.Product.FindAll(db.Cond{
		"id":     productIDs,
		"status": data.ProductStatusApproved,
	})

	lightningCollection.Products = NewProductList(ctx, products)
	return lightningCollection
}
