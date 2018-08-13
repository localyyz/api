package sync

import (
	"context"
	"log"

	"upper.io/db.v3"

	"fmt"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

func ShopifyCollectionListingsRemove(ctx context.Context) error {
	list := ctx.Value("sync.list").([]*shopify.CollectionList)

	// putting all the external ids into one list
	var collectionIDs []int64
	for _, c := range list {
		collectionIDs = append(collectionIDs, c.ID)
	}

	// get all the collections
	collections, err := data.DB.Collection.FindAll(db.Cond{"external_id": collectionIDs})
	if err != nil {
		return errors.Wrap(err, "Shopify syncer failed to find collections")
	}

	// run through all the collections
	for _, c := range collections {
		c.DeletedAt = data.GetTimeUTCPointer()
		err := data.DB.Collection.Save(c)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Shopiy syncer failed to mark collection: %d as deleted", c.ID))
		}
	}

	return nil
}

func ShopifyCollectionListingsUpdate(ctx context.Context) error {
	list := ctx.Value("sync.list").([]*shopify.CollectionList)
	client := ctx.Value("shopify.client").(*shopify.Client)

	for _, c := range list {

		// getting the merchant collection from db
		mC, err := data.DB.Collection.FindOne(db.Cond{"external_id": c.ID})
		if err != nil && err != db.ErrNoMoreRows {
			return errors.Wrap(err, "Shopify syncer had some error reading the db while attempting to update collection")
		}

		// check if it doesnt exist -> create it
		if err == db.ErrNoMoreRows {
			err := ShopifyCollectionListingsCreate(ctx)
			if err != nil {
				return err
			}
			continue
		}

		// updating the title and image
		mC.Name = c.Title
		if c.Image != nil {
			mC.ImageURL = c.Image.Src
		}

		err = data.DB.Collection.Save(mC)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Error: Shopify syncer unable to save collection with id: %d ", mC.ID))
		}

		// getting the product IDs for the collection
		extIDs, resp, err := client.CollectionList.ListProductIDs(ctx, c.ID)
		if err != nil || resp.StatusCode != http.StatusOK {
			return err
		}

		// add new products
		products, err := data.DB.Product.FindAll(db.Cond{"external_id": extIDs})
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Error: Shopify syncer unable to find products"))
		}

		for _, product := range products {
			AddProductToCollection(product.ID, mC.ID)
		}

		// not concerned about removing products due to use of table as metadata
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
			Featured:    false,
			MerchantID:  place.ID,
			ExternalID:  &coll.ID,
		}

		if coll.Image != nil {
			collection.ImageURL = coll.Image.Src
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

		// finding all the products
		products, err := data.DB.Product.FindAll(db.Cond{"external_id": extIDs})
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Shopify could not find product in db for collection: %d", collection.ID))
		}

		// add to collection_products
		for _, p := range products {
			cp := data.CollectionProduct{ProductID: p.ID, CollectionID: collection.ID}
			err = data.DB.CollectionProduct.Create(cp)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("Shopify could not save to collection_products for collection: %d", collection.ID))
			}
		}
	}
	return nil
}

func AddProductToCollection(productID, collectionID int64) error {
	// add to collection_products
	cp := data.CollectionProduct{
		ProductID:    productID,
		CollectionID: collectionID,
	}
	err := data.DB.CollectionProduct.Create(cp)
	if err != nil {
		if pErr, ok := err.(*pq.Error); ok {
			log.Println(pErr.Code.Name())
			if pErr.Code.Name() != "unique_violation" {
				return err
			}
			return nil
		}
		return errors.Wrapf(err, "Shopify syncer could not save to collection_products for collection: %d", collectionID)
	}
	return nil
}
