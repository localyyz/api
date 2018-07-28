package scheduler

import (
	"context"
	"net/http"
	"net/url"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"upper.io/db.v3"

	"time"

	"github.com/pressly/lg"
)

const LocalyyzStoreId = 4164
const DotdCollectionId = 76596346998

func (h *Handler) SyncDOTD() {
	h.wg.Add(1)
	defer h.wg.Done()

	s := time.Now()
	lg.Info("job_sync_deals running...")
	defer func() {
		lg.Infof("job_sync_deals finished in %s", time.Since(s))
	}()

	ctx := context.Background()
	// getting the shopify cred
	cred, err := data.DB.ShopifyCred.FindOne(db.Cond{"place_id": LocalyyzStoreId})
	if err != nil {
		lg.Alert("Sync DOTD: Failed to get Shopify Credentials")
		return
	}

	// creating the client
	client := shopify.NewClient(nil, cred.AccessToken)
	client.BaseURL, err = url.Parse(cred.ApiURL)
	if err != nil {
		lg.Alert("Sync DOTD: Failed to instantiate Shopify client")
		return
	}

	// get the dotd collection
	collection, resp, err := client.CollectionList.Get(ctx, DotdCollectionId)
	if err != nil || resp.StatusCode != http.StatusOK {
		lg.Alert("Sync DOTD: Failed to get collection from Shopify")
		return
	}

	// getting the product ids
	extIDs, resp, err := client.CollectionList.ListProductIDs(ctx, collection.ID)
	if err != nil || resp.StatusCode != http.StatusOK {
		lg.Alert("Sync DOTD: Failed to get product IDs of the collection")
		return
	}

	// get the queued collections
	var dotdColl []*data.Collection
	res := data.DB.Collection.Find(
		db.Cond{
			"lightning": true,
			"status":    data.CollectionStatusQueued,
		},
	).Limit(1).OrderBy("end_at DESC")
	res.All(&dotdColl)

	// setting the initial time
	// collections are scheduled +1 day from the end time of the latest upcoming deal in the db
	// if no upcoming deal schedule +1 day from TODAY
	now := time.Now()
	lastEndTime := &now
	if len(dotdColl) > 0 {
		lastEndTime = dotdColl[0].EndAt
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
				lg.Alert("Sync DOTD: Failed to retrieve product from DB")
				return
			}
		}

		// check if the product already exists as part of a collection
		var cp *data.CollectionProduct
		res := data.DB.CollectionProduct.Find(db.Cond{"product_id": product.ID})
		err = res.One(&cp)
		if err != nil && err != db.ErrNoMoreRows {
			lg.Alert("Sync DOTD: Failed to retrieve entry from collection_products")
			return

		}

		// did not find any entry -> create new collection
		if err == db.ErrNoMoreRows {
			// get product variant
			// A DOTD PRODUCT CANNOT BE USED IN MORE THAN ONE COLLECTION
			var pv *data.ProductVariant
			res = data.DB.ProductVariant.Find(db.Cond{"product_id": product.ID, "limits": db.Gt(0)})
			err = res.One(&pv)
			if err != nil {
				lg.Alert("Sync DOTD: Failed to retrieve product variant from DB")
				return
			}

			// get product image
			var pImg *data.ProductImage
			res = data.DB.ProductImage.Find(db.Cond{"product_id": product.ID, "ordering": 1})
			err = res.One(&pImg)
			if err != nil {
				lg.Alert("Sync DOTD: Failed to retrieve product image from DB")
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
			// manually setting the time to 12 o'clock and adding +saveCount days to ensure deal is scheduled at noon of the next day of the last deal
			start := time.Date(lastEndTime.Year(), lastEndTime.Month(), lastEndTime.Day()+saveCount, 16, 0, 0, 0, time.UTC)
			end := start.Add(time.Hour)

			newColl.StartAt = &start
			newColl.EndAt = &end

			// saving the new collection
			err = data.DB.Collection.Save(newColl)
			if err != nil {
				lg.Alert("Sync DOTD: Failed to save new DOTD collection to DB")
				return
			}

			// create the entry in the collection_products table
			newCollProd := data.CollectionProduct{ProductID: product.ID, CollectionID: newColl.ID}
			err = data.DB.CollectionProduct.Create(newCollProd)
			if err != nil {
				lg.Alert("Sync DOTD: Failed to save entry in collection_products")
				return
			}

			// increment the number of collections saved
			saveCount++
		}
	}
}
