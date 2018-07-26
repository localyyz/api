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

	// get the collections from the store
	collections, response, err := client.CustomCollection.Get(ctx, nil)
	if err != nil || response.StatusCode != http.StatusOK {
		return
	}

	// iterating through collections
	for _, collection := range collections {
		if collection.Title == "DOTD" {
			// getting all the product ids associated with the collection
			productIDs, resp, err := client.CollectionList.ListProductIDs(ctx, collection.ID)
			if err != nil || resp.StatusCode != http.StatusOK {
				return
			}
			for _, externalProductID := range productIDs {
				// find if the product is synced
				product, err := data.DB.Product.FindByExternalID(externalProductID)
				if err != nil {
					if err == db.ErrNoMoreRows {
						//if not found just skip the rest of the lines
						continue
					} else {
						// some other error
						return
					}
				}

				// see if the entry exists in the collection_products table
				var collectionProduct *data.CollectionProduct
				res := data.DB.CollectionProduct.Find(db.Cond{"product_id": product.ID})
				err = res.One(&collectionProduct)
				if err != nil {
					// the entry does not exist - if it exists then skip
					if err == db.ErrNoMoreRows {
						// get the variants
						productVariant, err := data.DB.ProductVariant.FindByID(product.ID)
						if err != nil {
							return
						}
						// get the images
						productImage, err := data.DB.ProductImage.FindByID(product.ID)
						if err != nil {
							return
						}
						// get the upcoming collections
						var dotdCollection []*data.Collection
						res := data.DB.Collection.Find(
							db.Cond{
								"lightning": true,
								"status":    data.CollectionStatusQueued,
							},
						).OrderBy("end_at DESC")
						err = res.All(&dotdCollection)
						if err != nil {
							return
						}

						// creating the new collections
						newCollection := &data.Collection{}
						newCollection.Name = product.Title
						newCollection.Description = product.Description
						newCollection.ImageURL = productImage.ImageURL
						newCollection.ImageWidth = productImage.Width
						newCollection.ImageHeight = productImage.Height
						newCollection.Gender = product.Gender
						newCollection.Lightning = true
						if len(dotdCollection) > 0 {
							// there are upcoming dotds
							lastStartTime := dotdCollection[0].StartAt
							lastEndTime := dotdCollection[0].EndAt
							startTime := lastStartTime.AddDate(0, 0, 1)
							endTime := lastEndTime.AddDate(0, 0, 1)
							newCollection.StartAt = &startTime
							newCollection.EndAt = &endTime
						} else {
							// there are no upcoming dotds
							now := time.Now()
							if now.Hour() < 12 {
								// schedule for today
								startTime := time.Date(now.Year(), now.Month(), now.Day(), 16, 0, 0, 0, time.UTC)
								endTime := startTime.Add(time.Hour)
								newCollection.StartAt = &startTime
								newCollection.EndAt = &endTime
							} else {
								// schedule for tomorrow
								startTime := time.Date(now.Year(), now.Month(), now.Day()+1, 16, 0, 0, 0, time.UTC)
								endTime := startTime.Add(time.Hour)
								newCollection.StartAt = &startTime
								newCollection.EndAt = &endTime
							}
						}
						newCollection.Status = data.CollectionStatusQueued
						newCollection.Cap = productVariant.Limits / 10 //10% of the original
						// saving the new collection
						err = data.DB.Collection.Save(newCollection)
						if err != nil {
							return
						}
						// creating the collection product - we need the ID from the db
						savedCollection, err := data.DB.Collection.FindOne(db.Cond{"name": product.Title})
						if err != nil {
							return
						}
						newCollectionProduct := data.CollectionProduct{ProductID: product.ID, CollectionID: savedCollection.ID}
						err = data.DB.CollectionProduct.Create(newCollectionProduct)
						if err != nil {
							return
						}
					} else {
						//some other error
						return
					}
				}
			}
		}
	}
}
