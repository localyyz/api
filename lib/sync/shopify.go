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
		// TODO: put this back in in the future. for now, sync everytime
		//if product.Status == data.ProductStatusRejected {
		//return ErrProductRejected
		//}
		lg.SetEntryField(ctx, "product_id", product.ID)

		product.Status = data.ProductStatusProcessing
		// lock product in as processing
		if err := data.DB.Product.Save(product); err != nil {
			return errors.Wrap(err, "failed to lock product for update")
		}

		listenerI := ctx.Value(SyncListenerCtxKey)

		// async syncing of variants / product images
		// NOTE: this go func should have nothing to do with the context.
		go func() {
			if listener, ok := listenerI.(Listener); ok {
				// inform caller that we're done
				defer func() { listener <- 1 }()
			}
			syncer := &productSyncer{
				place:   place,
				product: product,
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
		}()
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

		// pull the caches out of context.
		categoryCache := ctx.Value("category.cache").(map[string]*data.Category)
		blacklistCache := ctx.Value("category.blacklist").(map[string]*data.Blacklist)
		listenerI := ctx.Value(SyncListenerCtxKey)

		// async syncing of variants / product images
		// NOTE: this go func should have nothing to do with the context.
		go func() {
			if listener, ok := listenerI.(Listener); ok {
				// inform caller that we're done
				defer func() { listener <- 1 }()
			}

			// create syncer product scope
			syncer := &productSyncer{
				place:          place,
				product:        product,
				categoryCache:  categoryCache,
				blacklistCache: blacklistCache,
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
						return
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

			if err := syncer.SyncVariants(p.Variants); err != nil {
				lg.Warnf("shopify add variant: %v", err)
				return
			}
			if err := syncer.SyncImages(p.Images); err != nil {
				lg.Warnf("shopify add images: %v", err)
				return
			}
			if err := syncer.SyncScore(); err != nil {
				lg.Warnf("shopify score%v", err)
				return
			}
			syncer.Finalize()
		}()
	}
	return nil
}
