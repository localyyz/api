package sync

import (
	"context"

	"upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"github.com/pkg/errors"
	"fmt"
	"net/http"
)


func ShopifyCollectionListingsRemove(ctx context.Context) error {
	list := ctx.Value("sync.list").([]*shopify.CollectionList)

	for _, c := range list {
		// find the collection in our db
		dbColl, err := data.DB.Collection.FindOne(db.Cond{"external_id": c.ID})
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Shopify syncer failed to find collection with external id: %d", c.ID))
		}

		// marking the collection as deleted
		dbColl.Status = data.CollectionStatusDeleted
		err = data.DB.Collection.Save(dbColl)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Shopiy syncer failed to delete collection: %d", dbColl.ID))
		}
	}
	return nil
}

func ShopifyCollectionListingsUpdate(ctx context.Context) error {
	list := ctx.Value("sync.list").([]*shopify.CollectionList)
	client := ctx.Value("shopify.client").(*shopify.Client)

	for _, c := range list {
		// check if it already exists
		if exists, _ := data.DB.Collection.Find(db.Cond{"external_id": c.ID}).Exists(); !exists {
			err := ShopifyCollectionListingsCreate(ctx)
			if err != nil {
				return err
			}
			continue
		}

		// perform the update
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
			// its not a part of collection_products
			if exist, _ := data.DB.CollectionProduct.Find(db.Cond{"product_id": product.ID}).Exists(); !exist {
				// add to collection_products
				cp := data.CollectionProduct{ProductID: product.ID, CollectionID: mC.ID}
				err = data.DB.CollectionProduct.Create(cp)
				if err != nil {
					return errors.Wrap(err, fmt.Sprintf("Shopify syncer could not save to collection_products for collection: %d", mC.ID))
				}
			}
		}

		// remove products from the collection
		prodColls, _ := data.DB.CollectionProduct.FindByCollectionID(mC.ID)
		for _, prodColl := range prodColls {
			product, err := data.DB.Product.FindByID(prodColl.ProductID)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("Shopify syncer could not find product with id: %d", product.ID))
			}
			if !containsProduct(product.ExternalID, extIDs) {
				err := data.DB.CollectionProduct.Delete(prodColl)
				if err != nil {
					return errors.Wrap(err, fmt.Sprintf("Shopify syncer failed to remove from product collection for collection: %d", mC.ID))
				}
			}
		}

	}
	return nil
}

func ShopifyCollectionListingsCreate(ctx context.Context) error {
	place := ctx.Value("sync.place").(*data.Place)
	list := ctx.Value("sync.list").([]*shopify.CollectionList)
	client := ctx.Value("shopify.client").(*shopify.Client)

	for _, coll := range list {
		//validate collection does not exist
		if exist, err := data.DB.Collection.Find(db.Cond{"external_id": coll.ID}).Exists(); exist {
			return errors.Wrap(err, "Shopify collection already exists")
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

		// getting the product ids
		extIDs, resp, err := client.CollectionList.ListProductIDs(ctx, coll.ID)
		if err != nil || resp.StatusCode != http.StatusOK {
			return errors.Wrap(err, fmt.Sprintf("Shopify could not get the product ids for collection : %d", collection.ID))
		}

		for _, extID := range extIDs {
			// validate product exists
			if exist, _ := data.DB.Product.Find(db.Cond{"external_id": extID}).Exists(); !exist {
				// product does not exist skip
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

func containsProduct(productID *int64, list []int64) bool {
	for _, p := range list {
		if *productID == p {
			return true
		}
	}
	return false
}
