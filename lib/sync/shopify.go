package sync

import (
	"context"
	"net/url"
	"strconv"
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
		err := data.DB.Product.Find(db.Cond{
			"place_id":    place.ID,
			"external_id": p.ProductID,
		}).Delete()
		if err != nil {
			lg.Warnf("failed to delete product %s with %+v", p.Handle, err)
			return nil
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
				// skip entire update if product is not found

				// NOTE: this happens when shopify webhook calls
				// comes through out-of-order. some time receives
				// update before create. For now, ignore and silently fail
				continue
			}
			return errors.Wrap(err, "failed to fetch product")
		}
		lg.SetEntryField(ctx, "product_id", product.ID)

		// if product image is empty
		if product.ImageUrl == "" {
			// parse product images
			for _, img := range p.Images {
				imgUrl, _ := url.Parse(img.Src)
				imgUrl.Scheme = "https"
				if img.Position == 1 {
					product.ImageUrl = imgUrl.String()
				}
				// TODO: -> product_images table and reference a variant
				product.Etc.Images = append(product.Etc.Images, imgUrl.String())
			}
			// save product
			data.DB.Product.Save(product)
		}

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

			dbVariant.Etc.Price, _ = strconv.ParseFloat(v.Price, 64)
			dbVariant.Etc.PrevPrice, _ = strconv.ParseFloat(v.CompareAtPrice, 64)
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
			Etc:            data.ProductEtc{},
		}

		// parse product images if adding for the first time
		for _, img := range p.Images {
			imgUrl, _ := url.Parse(img.Src)
			imgUrl.Scheme = "https"
			if img.Position == 1 {
				product.ImageUrl = imgUrl.String()
			}
			// TODO: -> product_images table and reference a variant
			product.Etc.Images = append(product.Etc.Images, imgUrl.String())
		}

		// find product category + gender
		parsedData := ParseProduct(ctx, p.Title, p.Tags, p.ProductType)
		product.Gender = parsedData.Gender
		if len(parsedData.Value) > 0 {
			product.Category = data.ProductCategory{Type: parsedData.Type, Value: parsedData.Value}
		}

		// save product to database. Exit if fail
		if err := data.DB.Product.Save(product); err != nil {
			return errors.Wrap(err, "failed to save product")
		}
		if len(parsedData.Value) > 0 {
			// TODO: move product category onto product table
			data.DB.ProductTag.Save(&data.ProductTag{
				PlaceID:   place.ID,
				ProductID: product.ID,
				Type:      data.ProductTagTypeCategory,
				Value:     parsedData.Value,
			})
		} else {
			lg.Warnf("parseTags: pl(%s) pr(%s) gnd(%v) without category", place.Name, p.Handle, product.Gender)
		}

		lg.SetEntryField(ctx, "product_id", product.ID)

		// bulk insert variants
		q := data.DB.InsertInto("product_variants").
			Columns("product_id", "place_id", "limits", "description", "offer_id", "etc")
		b := q.Batch(len(p.Variants))
		go func() {
			defer b.Done()
			for _, v := range p.Variants {
				price, _ := strconv.ParseFloat(v.Price, 64)

				prevPrice, _ := strconv.ParseFloat(v.CompareAtPrice, 64)
				variantEtc := data.ProductVariantEtc{
					Price:     price,
					PrevPrice: prevPrice,
					Sku:       v.Sku,
				}

				// variant option values
				for _, o := range v.OptionValues {
					vv := strings.ToLower(o.Value)
					switch strings.ToLower(o.Name) {
					case "size":
						variantEtc.Size = vv
					case "color":
						variantEtc.Color = vv
					default:
						// pass
					}
				}

				b.Values(product.ID, place.ID, v.InventoryQuantity, v.Title, v.ID, variantEtc)
			}
		}()
		if err := b.Wait(); err != nil {
			return errors.Wrap(err, "failed to create product variants")
		}

	}
	return nil
}
