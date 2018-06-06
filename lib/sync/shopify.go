package sync

import (
	"context"
	"net/http"
	"strings"

	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/htmlx"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"github.com/pkg/errors"
	"github.com/pressly/lg"
)

var (
	ErrOutofStock = errors.New("out of stock")
)

func ShopifyProductListingsRemove(ctx context.Context) error {
	list := ctx.Value("sync.list").([]*shopify.ProductList)
	place := ctx.Value("sync.place").(*data.Place)

	for _, p := range list {
		// check if product already exists in our system
		dbProduct, err := data.DB.Product.FindOne(db.Cond{
			"place_id":    place.ID,
			"external_id": p.ProductID,
			"deleted_at":  db.IsNotNull(),
		})
		if err != nil {
			lg.Warnf("failed to delete product %s with %+v", p.Handle, err)
			return nil
		}
		// Mark as deleted at and save
		dbProduct.DeletedAt = data.GetTimeUTCPointer()
		data.DB.Product.Save(dbProduct)
	}

	return nil
}

func setVariants(p *data.Product, variants ...*shopify.ProductVariant) error {
	// bulk insert variants
	q := data.DB.InsertInto("product_variants").
		Columns("product_id", "place_id", "limits", "description", "offer_id", "price", "prev_price", "etc")
	b := q.Batch(len(variants))
	go func() {
		defer b.Done()
		for i, v := range variants {
			price, prevPrice := setPrices(
				v.Price,
				v.CompareAtPrice,
			)

			if i == 0 {
				p.Price = price
				if price > 0 && prevPrice > price {
					p.DiscountPct = pctRound(price/prevPrice, 1)
				}
			}

			etc := data.ProductVariantEtc{Sku: v.Sku}
			// variant option values
			for _, o := range v.OptionValues {
				vv := strings.ToLower(o.Value)
				switch strings.ToLower(o.Name) {
				case "size":
					etc.Size = vv
				case "color":
					etc.Color = vv
				default:
					// pass
				}
			}
			b.Values(p.ID, p.PlaceID, v.InventoryQuantity, v.Title, v.ID, price, prevPrice, etc)
		}
	}()
	if err := b.Wait(); err != nil {
		return errors.Wrap(err, "failed to create product variants")
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

		// Update the vendor
		product.Brand = p.Vendor

		// iterate product variants and update quantity limit
		isOutofStock := true
		for i, v := range p.Variants {
			dbVariant, err := data.DB.ProductVariant.FindByOfferID(v.ID)
			if err != nil {
				if err != db.ErrNoMoreRows {
					continue
				}
				// create new db variant previously unavailable
				dbVariant = &data.ProductVariant{
					PlaceID:   place.ID,
					ProductID: product.ID,
					OfferID:   v.ID,
					Etc:       data.ProductVariantEtc{},
				}
			}

			// set price/compare price
			dbVariant.Price, dbVariant.PrevPrice = setPrices(
				v.Price,
				v.CompareAtPrice,
			)

			if i == 0 {
				product.Price = dbVariant.Price
				if dbVariant.Price > 0 && dbVariant.PrevPrice > dbVariant.Price {
					product.DiscountPct = pctRound(dbVariant.Price/dbVariant.PrevPrice, 1)
				}
			}

			for _, o := range v.OptionValues {
				vv := strings.ToLower(o.Value)
				switch strings.ToLower(o.Name) {
				case "size":
					dbVariant.Etc.Size = vv
				case "color":
					dbVariant.Etc.Color = vv
				default:
					// pass
				}
			}
			dbVariant.Limits = int64(v.InventoryQuantity)
			dbVariant.Description = strings.ToLower(v.Title)
			if err := data.DB.ProductVariant.Save(dbVariant); err != nil {
				return errors.Wrap(err, "failed to update product variant")
			}

			// if any variant is in stock, set isOutofStock to false
			if isOutofStock && dbVariant.Limits > 0 {
				isOutofStock = false
			}
		}

		// update product images if product image is empty
		err = setImages(&shopifyImageSyncer{Product: product, Client: &http.Client{}}, p.Images...)

		// should make this more testable
		//
		// TODO: handle error wrapper?
		if isOutofStock && err == nil {
			err = ErrOutofStock
		}

		// update product status
		product.Status = finalizeStatus(
			ctx,
			len(product.Category.Value) > 0,
			err,
			p.Title,
			p.Tags,
			p.ProductType,
		)

		err = ScoreProduct(&shopifyImageScorer{Product: product, Place: place})
		if err != nil {
			lg.Warnf("Error: could not set score for product id: %d", product.ID)
		}

		//save
		data.DB.Product.Save(product)

	}

	return nil
}

func finalizeStatus(ctx context.Context, hasCategory bool, err error, inputs ...string) data.ProductStatus {

	if err != nil {
		return data.ProductStatusRejected
	}

	// if blacklisted, demote the product status
	if SearchBlackList(ctx, inputs...) {
		if hasCategory {
			// mark as pending, blacklisted but found a category
			return data.ProductStatusPending
		} else {
			// reject if we did not find a category
			return data.ProductStatusRejected
		}
	}

	if hasCategory {
		return data.ProductStatusApproved
	}
	return data.ProductStatusPending
}

func ShopifyProductListingsCreate(ctx context.Context) error {
	place := ctx.Value("sync.place").(*data.Place)
	list := ctx.Value("sync.list").([]*shopify.ProductList)

	for _, p := range list {
		if !p.Available {
			// Skip any product _not_ available
			continue
		}

		product := &data.Product{
			PlaceID:        place.ID,
			ExternalID:     &p.ProductID,
			ExternalHandle: p.Handle,
			Title:          p.Title,
			Description:    htmlx.CaptionizeHtmlBody(p.BodyHTML, -1),
			Brand:          p.Vendor,
			Status:         data.ProductStatusPending,
		}

		// find product category + gender
		parsedData := ParseProduct(ctx, p.Title, p.Tags, p.ProductType)
		product.Gender = parsedData.Gender

		if len(parsedData.Value) > 0 {
			product.Category = data.ProductCategory{
				Type:  parsedData.Type,
				Value: parsedData.Value,
			}
		}

		lg.SetEntryField(ctx, "product_id", product.ID)

		// product variants
		setVariants(product, p.Variants...)

		// set imgs
		err := setImages(&shopifyImageSyncer{Product: product}, p.Images...)

		// product status
		product.Status = finalizeStatus(
			ctx,
			len(product.Category.Value) > 0,
			err,
			p.Title,
			p.Tags,
			p.ProductType,
		)

		err = ScoreProduct(&shopifyImageScorer{Product: product, Place: place})
		if err != nil {
			lg.Warnf("Error: could not set score for product id: %d", product.ID)
		}

		// save
		data.DB.Product.Save(product)
	}
	return nil
}
