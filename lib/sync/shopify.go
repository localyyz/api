package sync

import (
	"context"

	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/htmlx"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"fmt"
	"github.com/pkg/errors"
	"github.com/pressly/lg"
	"net/http"
	"net/url"
)

var (
	ErrOutofStock      = errors.New("out of stock")
	ErrProductExist    = errors.New("exists")
	ErrProductRejected = errors.New("rejected")
	ErrCollectionExist = errors.New("exists")
	SyncListenerCtxKey = "sync.listener"
)

type Listener chan int

func ShopifyProductListingsRemove(ctx context.Context) error {
	list := ctx.Value("sync.list").([]*shopify.ProductList)
	place := ctx.Value("sync.place").(*data.Place)
	for _, p := range list {
		// check if product already exists in our system
		dbProduct, err := data.DB.Product.FindOne(db.Cond{
			"place_id":    place.ID,
			"external_id": p.ProductID,
			"deleted_at":  db.IsNull(),
		})
		if err != nil {
			lg.Warnf("failed to delete product %s with %+v", p.Handle, err)
			return nil
		}
		lg.SetEntryField(ctx, "product_id", dbProduct.ID)
		// Mark as deleted at and save
		dbProduct.DeletedAt = data.GetTimeUTCPointer()
		dbProduct.Status = data.ProductStatusDeleted
		if err := data.DB.Product.Save(dbProduct); err != nil {
			return err
		}
	}
	return nil
}

func ShopifyProductListingsUpdate(ctx context.Context) error {
	place := ctx.Value("sync.place").(*data.Place)
	list := ctx.Value("sync.list").([]*shopify.ProductList)

	for _, p := range list {
		// load the product from database
		product, err := data.DB.Product.FindOne(db.Cond{
			"place_id":    place.ID,
			"external_id": p.ProductID, // externalID
		})
		if err != nil {
			if err == db.ErrNoMoreRows {
				// redirect to shopify products create
				return ShopifyProductListingsCreate(ctx)
				// NOTE: this happens when shopify webhook calls
				// comes through out-of-order. some time receives
				// update before create. For now, ignore and silently fail
			}
			return errors.Wrap(err, "failed to fetch product")
		}
		// TODO: put this back in in the future. for now, sync everytime
		//if product.Status == data.ProductStatusRejected {
		//return ErrProductRejected
		//}
		lg.SetEntryField(ctx, "product_id", product.ID)

		product.Status = data.ProductStatusProcessing
		data.DB.Product.Save(product) // lock product in as processing
		syncer := &productSyncer{
			place:   place,
			product: product,
		}

		// async syncing of variants / product images
		go func(ctx context.Context) {
			if listener, ok := ctx.Value(SyncListenerCtxKey).(Listener); ok {
				// inform caller that we're done
				defer func() { listener <- 1 }()
			}
			if err := syncer.SyncVariants(p.Variants); err != nil {
				lg.Warnf("shopify add variant: %v", err)
				return
			}
			if err := syncer.SyncImages(p.Images); err != nil {
				lg.Warnf("shopify add images: %v", err)
				return
			}
			if err := syncer.SyncScore(); err != nil {
				lg.Warnf("shopify score: %v", err)
				return
			}
			syncer.Finalize()
		}(ctx)
	}

	return nil
}

func ShopifyProductListingsCreate(ctx context.Context) error {
	place := ctx.Value("sync.place").(*data.Place)
	list := ctx.Value("sync.list").([]*shopify.ProductList)

	for _, p := range list {
		if !p.Available {
			// Skip any product _not_ available
			continue
		}

		// validate product does not exist
		if exist, _ := data.DB.Product.Find(db.Cond{"external_id": p.ProductID}).Exists(); exist {
			return ErrProductExist
		}

		product := &data.Product{
			PlaceID:        place.ID,
			ExternalID:     &p.ProductID,
			ExternalHandle: p.Handle,
			Title:          p.Title,
			Description:    htmlx.CaptionizeHtmlBody(p.BodyHTML, -1),
			Brand:          p.Vendor,
			Status:         data.ProductStatusProcessing,
		}
		syncer := &productSyncer{
			place:   place,
			product: product,
		}

		// find product category + gender
		parsedData, err := ParseProduct(ctx, p.Title, p.Tags, p.ProductType)
		if err != nil {
			// see parse product comment for logic on blacklisting product
			// throw away the product. and continue on
			// TODO: keep track of how many products are rejected

			//  create a logic map
			//   - x blacklist + o category -> reject
			//   - x blacklist + x category -> pending?
			//   - o blacklist + o category -> pending?
			//   - o blacklist + x category -> good
			if err == ErrBlacklisted {
				if len(parsedData.Value) == 0 {
					syncer.FinalizeStatus(data.ProductStatusRejected)
					// no category and blacklisted -> return
					return err
				}
				// blacklisted but has category
				syncer.product.Status = data.ProductStatusPending
			}
		}
		if len(parsedData.Value) == 0 {
			// not blacklisted but no category
			syncer.product.Status = data.ProductStatusPending
		}
		syncer.product.Gender = parsedData.Gender
		syncer.product.Category = data.ProductCategory{
			Type:  parsedData.Type,
			Value: parsedData.Value,
		}

		// must save product before moving on to other contexts
		if err := data.DB.Product.Save(syncer.product); err != nil {
			return errors.Wrap(err, "shopify product create")
		}
		lg.SetEntryField(ctx, "product_id", syncer.product.ID)

		// async syncing of variants / product images
		go func(s *productSyncer) {
			if listener, ok := ctx.Value(SyncListenerCtxKey).(Listener); ok {
				// inform caller that we're done
				defer func() { listener <- 1 }()
			}
			if err := s.SyncVariants(p.Variants); err != nil {
				lg.Warnf("shopify add variant: %v", err)
				return
			}
			if err := s.SyncImages(p.Images); err != nil {
				lg.Warnf("shopify add images: %v", err)
				return
			}
			if err := s.SyncScore(); err != nil {
				lg.Warnf("shopify score%v", err)
				return
			}

			s.Finalize()
		}(syncer)
	}
	return nil
}

func ShopifyCollectionListingsRemove(ctx context.Context) error {
	list := ctx.Value("sync.list").([]*shopify.CollectionList)

	for _, c := range list {
		// find the collection in our db
		dbColl, err := data.DB.Collection.FindOne(db.Cond{"external_id": c.ID})
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Shopify syncer failed to find collection with external id: %d", c.ID))
		}

		// find all the collection products
		collProd, err := data.DB.CollectionProduct.FindAll(db.Cond{"collection_id": dbColl.ID})
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Shopify syncer failed to collection products for collection with external id: %d", c.ID))
		}

		// deleting all the collection products
		for _, cp := range collProd {
			err := data.DB.CollectionProduct.Delete(cp)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("Shopify syncer failed to delete collection product: %d for collection with external id: %d",cp.ProductID, c.ID))
			}
		}

		// deleting the collection
		err = data.DB.Collection.Delete(dbColl)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Shopiy syncer failed to delete collection with ID: %d", dbColl.ID))
		}
	}
	return nil
}

func ShopifyCollectionListingsUpdate(ctx context.Context) error {
	place := ctx.Value("sync.place").(*data.Place)
	list := ctx.Value("sync.list").([]*shopify.CollectionList)

	client, err := getShopifyClient(place.ID)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Shopify could not create client for place: %d", place.ID))
	}

	for _, c := range list {
		// check if it already exists
		if exists, _ := data.DB.Collection.Find(db.Cond{"external_id":c.ID}).Exists(); !exists {
			continue
		}

		// perform the update
		for _, c := range list {
			var mC *data.Collection
			err := data.DB.Collection.Find(db.Cond{"external_id": c.ID}).All(&mC)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("Error: Shopify syncer update collection listing but could not find collection with external id: %d", c.ID))
			}

			mC.Name = c.Title
			mC.ImageURL = c.Image.Src

			err = data.DB.Collection.Save(mC)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("Error: Shopify syncer unable to save collection with id: %d ", mC.ID))
			}

			// getting the product IDs
			extIDs, resp, err := client.CollectionList.ListProductIDs(ctx, c.ID)
			if err != nil || resp.StatusCode != http.StatusOK {
				return err
			}

			// add new products
			for _, extID := range extIDs {
				product, _ := data.DB.Product.FindByExternalID(extID)
				if exist, _ := data.DB.CollectionProduct.Find(db.Cond{"product_id":product.ID}).Exists(); !exist {
					// add to collection_products
					cp := data.CollectionProduct{ProductID: product.ID, CollectionID: mC.ID}
					err = data.DB.CollectionProduct.Create(cp)
					if err != nil {
						return errors.Wrap(err, fmt.Sprintf("Shopify could not save to collection_products for collection: %d", mC.ID))
					}
				}
			}

			// remove products from the collection
			prodColls, _ := data.DB.CollectionProduct.FindByCollectionID(mC.ID)
			for _, prodColl := range prodColls {
				if !containsProduct(prodColl.ProductID, extIDs) {
					err := data.DB.CollectionProduct.Delete(prodColl)
					if err != nil {
						return errors.Wrap(err, fmt.Sprintf("Shopify syncer failed to remove from product collection for collection: %d", mC.ID))
					}
				}
			}
		}
	}
	return nil
}

func ShopifyCollectionListingsCreate(ctx context.Context) error {
	place := ctx.Value("sync.place").(*data.Place)
	list := ctx.Value("sync.list").([]*shopify.CollectionList)

	for _, coll := range list {
		//validate collection does not exist
		if exist, _ := data.DB.Collection.Find(db.Cond{"external_id": coll.ID}).Exists(); exist {
			return ErrCollectionExist
		}

		// the new collection to save
		collection := data.Collection{
			Name:        coll.Title,
			Description: coll.BodyHTML,
			ImageURL:    coll.Image.Src,
			Featured:    false,
			MerchantID:  place.ID,
			Lightning:   false,
			ExternalID:  &coll.ID,
		}

		// saving the new collection
		err := data.DB.Collection.Save(collection)
		if err != nil {
			return errors.Wrap(err, "Shopify collection create")
		}

		client, err := getShopifyClient(place.ID)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Shopify could not create client for place: %d", place.ID))
		}

		// getting the product ids
		extIDs, resp, err := client.CollectionList.ListProductIDs(ctx, coll.ID)
		if err != nil || resp.StatusCode != http.StatusOK {
			return errors.Wrap(err, fmt.Sprintf("Shopify could not product ids for collection : %d", collection.ID))
		}

		for _, extID := range extIDs {
			// validate product exists
			if exist, _ := data.DB.Product.Find(db.Cond{"external_id": extID}).Exists(); !exist {
				continue
			}

			// retrieve the product
			p, err := data.DB.Product.FindByExternalID(extID)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("Shopify could not find product in db for collection: %d", collection.ID))
			}

			// add to collection_products
			cp := data.CollectionProduct{ProductID: p.ID, CollectionID: collection.ID}
			err = data.DB.CollectionProduct.Create(cp)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("Shopify could not save to collection_products for collection: %d", collection.ID))
			}
		}
	}
	return nil
}

func getShopifyClient(placeId int64) (*shopify.Client, error){
	// getting the shopify cred
	cred, err := data.DB.ShopifyCred.FindOne(db.Cond{"place_id": placeId})
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Shopify credentials for place_id: %d", placeId))
	}

	// creating the client
	client := shopify.NewClient(nil, cred.AccessToken)
	client.BaseURL, err = url.Parse(cred.ApiURL)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Shopify could not create client for place: %d", placeId))
	}

	return client, nil
}

func containsProduct(productID int64, list []int64) bool{
	for _, p := range list {
		if productID == p {
			return true
		}
	}
	return false
}