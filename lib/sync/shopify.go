package sync

import (
	"context"

	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/htmlx"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"github.com/pkg/errors"
	"github.com/pressly/lg"
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
		lg.SetEntryField(ctx, "product_id", product.ID)
		// if the product is rejected, keep is as rejected
		if product.Status == data.ProductStatusRejected {
			return ErrProductRejected
		}

		product.Status = data.ProductStatusProcessing
		// lock product in as processing
		if err := data.DB.Product.Save(product); err != nil {
			return errors.Wrap(err, "failed to lock product for update")
		}

		sync, _ := NewSyncer(ctx, product, place)
		if err := sync.Sync(p); err != nil {
			lg.Warnf("err: %+v", err)
		}
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
		// save product as processing for later consumption
		if err := data.DB.Product.Save(product); err != nil {
			return errors.Wrap(err, "shopify product create")
		}
		lg.SetEntryField(ctx, "product_id", product.ID)

		sync, _ := NewSyncer(ctx, product, place)
		if err := sync.Sync(p); err != nil {
			lg.Warnf("cause: %+v", err)
		}
	}
	return nil
}
