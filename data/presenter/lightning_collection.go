package presenter

import (
	"bitbucket.org/moodie-app/moodie-api/data"
	"context"
	"github.com/go-chi/render"
	"net/http"
	"upper.io/db.v3"
)

type LightningCollection struct {
	*data.Collection
	Product []*Product `json:"products"`
}

func (c *LightningCollection) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

/*
	Calculates the percentage complete -> counts the number of checkouts for the products in the collection
	Selects one product -> retrieves the first product from the collection which is in stock
*/
func PresentLightningCollection(ctx context.Context, collections []*data.Collection, isActive bool) ([]render.Renderer, error) {
	list := []render.Renderer{}
	for _, collection := range collections {

		lightningCollection := &LightningCollection{}
		//setting the collection
		lightningCollection.Collection = collection

		if isActive {
			// getting the product
			var lightningProduct data.Product
			var collectionProducts []*data.CollectionProduct
			err := data.DB.CollectionProduct.Find(db.Cond{"collection_id": collection.ID}).All(&collectionProducts)
			if err != nil {
				return nil, err
			}

			//iterate over the products from the collection
			for _, collectionProduct := range collectionProducts {
				var productVariant data.ProductVariant
				res := data.DB.ProductVariant.Find(db.Cond{"product_id": collectionProduct.ProductID})
				res.One(&productVariant)
				//the first variant
				if productVariant.Limits != 0 {
					res := data.DB.Product.Find(db.Cond{"id": collectionProduct.ProductID})
					res.One(&lightningProduct)
					break //return the first variant of the first product we find is in stock
				}
			}

			lightningCollection.Product = append(lightningCollection.Product, NewProduct(ctx, &lightningProduct))

		}

		//append to the final list to return
		list = append(list, lightningCollection)
	}

	return list, nil
}
