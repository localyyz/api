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

				// get the queued collections
				var dotdColl []*data.Collection
				res = data.DB.Collection.Find(
					db.Cond{
						"lightning": true,
						"status":    data.CollectionStatusQueued,
					},
				).OrderBy("end_at DESC")
				err = res.All(&dotdColl)
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

				if len(dotdColl) > 0 {
					// there are upcoming dotds -> so just schedule it the day after the last one
					lastStartTime := dotdColl[0].StartAt
					lastEndTime := dotdColl[0].EndAt
					startTime := lastStartTime.AddDate(0, 0, 1)
					endTime := lastEndTime.AddDate(0, 0, 1)
					newColl.StartAt = &startTime
					newColl.EndAt = &endTime
				} else {
					// no upcoming dotd -> schedule for tomorrow
					now := time.Now()
					startTime := time.Date(now.Year(), now.Month(), now.Day()+1, 16, 0, 0, 0, time.UTC)
					endTime := startTime.Add(time.Hour)
					newColl.StartAt = &startTime
					newColl.EndAt = &endTime
				}

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
			}
		}
	}
}
