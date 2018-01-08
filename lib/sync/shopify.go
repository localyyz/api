package sync

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	db "upper.io/db.v3"

	set "gopkg.in/fatih/set.v0"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/htmlx"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"github.com/gedex/inflector"
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
			lg.Alertf("failed to delete product %s with %+v", p.Handle, err)
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

		// iterate product variants and update quantity limit
		for _, v := range p.Variants {
			dbVariant, err := data.DB.ProductVariant.FindByOfferID(v.ID)
			if err != nil {
				lg.Alertf("variant offerID %d for product(%d) err: %+v", v.ID, product.ID, err)
				continue
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
		var variantPrice float64
		q := data.DB.InsertInto("product_variants").
			Columns("product_id", "place_id", "limits", "description", "offer_id", "etc")
		b := q.Batch(len(p.Variants))
		go func() {
			defer b.Done()
			for _, v := range p.Variants {
				price, _ := strconv.ParseFloat(v.Price, 64)
				variantPrice += price

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

		// bulk parse and insert product tags
		q = data.DB.InsertInto("product_tags").
			Columns("product_id", "place_id", "value", "type").
			Amend(func(query string) string {
				return query + ` ON CONFLICT DO NOTHING`
			})
		b = q.Batch(5)
		go func() {
			defer b.Done()

			// Flag if gender was found in any of the given fields
			var foundGender bool

			// Product Category
			if len(p.ProductType) > 0 {
				// Check if Gender can be found
				categorySplit := parseTags(p.ProductType)

				var (
					genderIdx   int
					foundGender bool
				)
				for i, c := range categorySplit {
					if gender := parseGender(c); gender != "" {
						b.Values(product.ID, place.ID, gender, data.ProductTagTypeGender)
						foundGender = true
						genderIdx = i
					}
				}
				if foundGender {
					// remove gender key from category
					categorySplit = append(categorySplit[:genderIdx], categorySplit[genderIdx+1:]...)
				}
				if len(categorySplit) > 0 {
					// Join the category back together into one
					b.Values(product.ID, place.ID, strings.Join(categorySplit, " "), data.ProductTagTypeCategory)
				}
			}

			// Product Vendor/Brand
			b.Values(product.ID, place.ID, strings.ToLower(p.Vendor), data.ProductTagTypeBrand)

			// Average variant prices
			b.Values(product.ID, place.ID, fmt.Sprintf("%.2f", variantPrice/float64(len(p.Variants))), data.ProductTagTypePrice)

			// Variant options (ie. Color, Size, Material)
			for _, o := range p.Options {
				var typ data.ProductTagType
				if err := typ.UnmarshalText([]byte(strings.ToLower(o.Name))); err != nil {
					continue
				}

				optSet := set.New()
				for _, v := range o.Values {
					vv := strings.ToLower(v)
					if optSet.Has(vv) {
						continue
					}
					b.Values(product.ID, place.ID, vv, typ)
					optSet.Add(vv)
				}
			}

			// Product tags for extra values
			for _, t := range parseTags(p.Tags) {
				if !foundGender {
					if gender := parseGender(t); gender != "" {
						b.Values(product.ID, place.ID, t, data.ProductTagTypeGender)
					}
				}
			}

			// Parse product title for potential gender data
			if !foundGender {
				for _, t := range parseTags(p.Title) {
					if gender := parseGender(t); gender != "" {
						b.Values(product.ID, place.ID, t, data.ProductTagTypeGender)
					}
				}
			}
		}()
		if err := b.Wait(); err != nil {
			return errors.Wrap(err, "failed to create product tags")
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

func parseGender(t string) string {
	if t == "man" || t == "woman" || t == "unisex" {
		// skip gender if specified, if we've already found gender
		return t
	}
	return ""
}

var tagRegex = regexp.MustCompile("[^a-zA-Z0-9-]+")

func hasNoLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func parseTags(tagStr string, optTags ...string) []string {
	tagStr = strings.ToLower(tagStr)
	tt := tagRegex.Split(tagStr, -1)

	tagSet := set.New()
	for _, t := range tt {
		if hasNoLetter(t) {
			// skip if not contain any alphanum letter
			continue
		}

		tt := inflector.Singularize(t)
		for {
			if tt == t {
				break
			}
			t = tt
			tt = inflector.Singularize(t)
		}
		if tt == "" {
			continue
		}
		tagSet.Add(t)
	}
	for _, t := range optTags {
		tagSet.Add(t)
	}

	return set.StringSlice(tagSet)
}
