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
	"github.com/sirupsen/logrus"
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
	lg.DefaultLogger = logrus.New()

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
			Gender:         data.ProductGender(place.Gender),
			Etc:            data.ProductEtc{},
			CreatedAt:      &p.CreatedAt,
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
		// save product to database. Exit if fail
		if err := data.DB.Product.Save(product); err != nil {
			return errors.Wrap(err, "failed to save product")
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

		if err := ShopifyProductTagsCreate(ctx, product, p); err != nil {
			return err
		}

	}
	return nil
}

func getProductCollections(ctx context.Context, productID int64) ([]*shopify.CustomCollection, error) {
	place := ctx.Value("sync.place").(*data.Place)

	cred, err := data.DB.ShopifyCred.FindByPlaceID(place.ID)
	if err != nil {
		return nil, err
	}

	cl := shopify.NewClient(nil, cred.AccessToken)
	cl.BaseURL, _ = url.Parse(cred.ApiURL)

	clist, _, err := cl.CustomCollection.Get(ctx, &shopify.CustomCollectionParam{ProductID: productID})
	if err != nil {
		return nil, err
	}

	return clist, nil
}
