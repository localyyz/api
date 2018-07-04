package presenter

import (
	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/go-chi/render"
	"math"
	"net/http"
	"upper.io/db.v3"
)

type LightningCollection struct {
	Collection         *data.Collection `json:"collection"`
	PercentageComplete float64          `json:"percentageComplete"`
	Product            *data.Product    `json:"product"`
}

func (c *LightningCollection) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

/*
	Calculates the percentage complete -> counts the number of checkouts for the products in the collection
	Selects one product -> retrieves the first product from the collection which is in stock
*/
func PresentActiveLightningCollection(collections []*data.Collection) ([]render.Renderer, error) {
	list := []render.Renderer{}
	for _, collection := range collections {
		lightningCollection := &LightningCollection{}

		//setting the collection
		lightningCollection.Collection = collection

		totalCheckouts := data.DB.Collection.GetCollectionCheckouts(collection)
		// better to be safe dividing by 0
		if collection.Cap != 0 {
			completionPercent := math.Round(float64(totalCheckouts)/float64(collection.Cap)*100) / 100
			lightningCollection.PercentageComplete = completionPercent
		}

		// getting the product
		var lightningProduct *data.Product
		var collectionProducts []*data.CollectionProduct
		err := data.DB.CollectionProduct.Find(db.Cond{"collection_id": collection.ID}).All(&collectionProducts)
		if err != nil {
			return nil, err
		}
		//iterate over the products from the collection
		for _, product := range collectionProducts {
			productVariant, err := data.DB.ProductVariant.FindByProductID(product.ProductID)
			if err != nil {
				return nil, err
			}
			//the first variant
			if productVariant[0].Limits != 0 {
				tempProduct, _ := data.DB.Product.FindByID(product.ProductID)
				lightningProduct = tempProduct
				break //return the first variant of the first product we find is in stock
			}
		}
		//append the product
		lightningCollection.Product = lightningProduct

		//append to the final list to return
		list = append(list, lightningCollection)
	}

	return list, nil
}

func PresentUpcomingLightningCollection(collections []*data.Collection) ([]render.Renderer, error){
	list := []render.Renderer{}
	for _, collection := range collections {
		lightningCollection := &LightningCollection{}

		//setting the collection
		lightningCollection.Collection = collection

		totalCheckouts := data.DB.Collection.GetCollectionCheckouts(collection)
		// better to be safe dividing by 0
		if collection.Cap != 0 {
			completionPercent := math.Round(float64(totalCheckouts)/float64(collection.Cap)*100) / 100
			lightningCollection.PercentageComplete = completionPercent
		}

		//append to the final list to return
		list = append(list, lightningCollection)
	}

	return list, nil
}