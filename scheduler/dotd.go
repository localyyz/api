package scheduler

import (
	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"context"
	"net/http"
	"net/url"
	"upper.io/db.v3"

	"time"
)

const LocalyyzStoreId = 4164
const DotdCollectionId = 76596346998

func (h *Handler) SyncDOTD() {
	ctx := context.Background()

	// getting the shopify cred
	cred, err := data.DB.ShopifyCred.FindOne(db.Cond{"place_id": LocalyyzStoreId})
	if err != nil {
		return
	}
	client := shopify.NewClient(nil, cred.AccessToken)
	client.BaseURL, err = url.Parse(cred.ApiURL)
	if err != nil {
		return
	}

	// get the dotd collection
	collection, resp, err := client.CollectionList.Get(ctx, DotdCollectionId)
	if err != nil || resp.StatusCode != http.StatusOK {
		return
	}

	// getting the product ids
	extIDs, resp, err := client.CollectionList.ListProductIDs(ctx, collection.ID)
	if err != nil || resp.StatusCode != http.StatusOK {
		return
	}

	// get the queued collections
	var dotdColl []*data.Collection
	res := data.DB.Collection.Find(
		db.Cond{
			"lightning": true,
			"status":    data.CollectionStatusQueued,
		},
	).OrderBy("end_at DESC")
	err = res.All(&dotdColl)
	if err != nil {
		return
	}

	// setting the initial time
	// collections are scheduled +1 day from startTime
	now := time.Now()
	startTime := &now
	if len(dotdColl) > 0 {
		startTime = dotdColl[0].StartAt
	}

	// used to keep track of how many new collections are being saved
	saveCount := 1

	// iterating through the product IDs
	for _, extID := range extIDs {
		// get the product
		product, err := data.DB.Product.FindByExternalID(extID)
		if err != nil {
			if err == db.ErrNoMoreRows {
				// product does not exist
				continue
			} else {
				return
			}
		}
		// check if the product already exists
		var cp *data.CollectionProduct
		res := data.DB.CollectionProduct.Find(db.Cond{"product_id": product.ID})
		err = res.One(&cp)
		if err != nil {
			if err == db.ErrNoMoreRows {
				// product does not exist -> create new collection

				// get product variant
				// A DOTD PRODUCT CANNOT BE USED IN MORE THAN ONE COLLECTION
				var pv *data.ProductVariant
				res := data.DB.ProductVariant.Find(db.Cond{"product_id": product.ID, "limits": db.Gt(0)})
				err := res.One(&pv)
				if err != nil {
					return
				}

				// get product image
				var pImg *data.ProductImage
				res = data.DB.ProductImage.Find(db.Cond{"product_id": product.ID, "ordering": 1})
				err = res.One(&pImg)
				if err != nil {
					return
				}

				// create the new collection
				newColl := &data.Collection{
					Name:        product.Title,
					Description: product.Description,
					ImageURL:    pImg.ImageURL,
					ImageWidth:  pImg.Width,
					ImageHeight: pImg.Height,
					Gender:      product.Gender,
					Lightning:   true,
					Status:      data.CollectionStatusQueued,
				}

				// calculating the cap
				cap := pv.Limits / 10
				if cap == 0 {
					newColl.Cap = 1
				} else {
					newColl.Cap = cap
				}

				// adding the start and end times
				start := time.Date(startTime.Year(), startTime.Month(), startTime.Day()+saveCount, 16, 0, 0, 0, time.UTC)
				end := start.Add(time.Hour)

				newColl.StartAt = &start
				newColl.EndAt = &end

				// saving the new collection
				err = data.DB.Collection.Save(newColl)
				if err != nil {
					return
				}

				// create the entry in the collection_products table
				newCollProd := data.CollectionProduct{ProductID: product.ID, CollectionID: newColl.ID}
				err = data.DB.CollectionProduct.Create(newCollProd)
				if err != nil {
					return
				}

				// increment the number of collections saved
				saveCount++
			}
		}
	}
}
