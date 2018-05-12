package sync

import (
	"context"
	"strings"

	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/htmlx"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"github.com/pkg/errors"
	"github.com/pressly/lg"
)

func ShopifyProductListingsRemove(ctx context.Context) error {
	list := ctx.Value("sync.list").([]*shopify.ProductList)
	place := ctx.Value("sync.place").(*data.Place)

	for _, p := range list {
		// check if product already exists in our system
		dbProduct, err := data.DB.Product.FindOne(db.Cond{
			"place_id":      place.ID,
			"external_id":   p.ProductID,
			"deleted_at <>": nil,
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
		for _, v := range variants {
			price, prevPrice := setPrices(
				v.Price,
				v.CompareAtPrice,
			)
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
				// skip entire update if product is not found

				// NOTE: this happens when shopify webhook calls
				// comes through out-of-order. some time receives
				// update before create. For now, ignore and silently fail
				continue
			}
			return errors.Wrap(err, "failed to fetch product")
		}
		lg.SetEntryField(ctx, "product_id", product.ID)

		// Update the vendor
		product.Brand = p.Vendor
		// update product images if product image is empty
		setImages(&shopifyImageSyncer{Product: product}, p.Images...)
		// Save
		data.DB.Product.Save(product)

		// iterate product variants and update quantity limit
		for _, v := range p.Variants {
			dbVariant, err := data.DB.ProductVariant.FindByOfferID(v.ID)
			if err != nil {
				if err != db.ErrNoMoreRows {
					lg.Warnf("variant offerID %d for product(%d) err: %+v", v.ID, product.ID, err)
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
		}
	}

	return nil
}

func finalizeStatus(ctx context.Context, hasCategory bool, inputs ...string) data.ProductStatus {
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
			Etc:            data.ProductEtc{},
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

		// check product blacklist
		product.Status = finalizeStatus(ctx, len(parsedData.Value) > 0, p.Title, p.Tags, p.ProductType)

		// save product to database. Exit if fail
		if err := data.DB.Product.Save(product); err != nil {
			return errors.Wrap(err, "failed to save product")
		}
		lg.SetEntryField(ctx, "product_id", product.ID)

		// product variants
		setVariants(product, p.Variants...)

		// set image and save
		setImages(&shopifyImageSyncer{Product: product}, p.Images...)
		// save
		data.DB.Product.Save(product)
	}
	return nil
}
