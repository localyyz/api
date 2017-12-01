package sync

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

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
		dbProduct, err := data.DB.Product.FindOne(db.Cond{
			"place_id":    place.ID,
			"external_id": p.ProductID,
		})
		if err != nil {
			return errors.Wrap(err, "failed to fetch product")
		}
		data.DB.Delete(dbProduct)
		lg.Alertf("product (%s) removed for place (%s)", p.Title, place.Name)
	}

	return nil
}

func ShopifyProductListings(ctx context.Context) error {
	list := ctx.Value("sync.list").([]*shopify.ProductList)
	place := ctx.Value("sync.place").(*data.Place)
	//whType := ctx.Value("sync.type").(shopify.Topic)

	for _, p := range list {
		product := &data.Product{
			PlaceID:        place.ID,
			ExternalID:     &p.ProductID,
			ExternalHandle: p.Handle,
			Title:          p.Title,
			Description:    htmlx.CaptionizeHtmlBody(p.BodyHTML, -1),
			Etc: data.ProductEtc{
				Brand: p.Vendor,
			},
		}

		// check if product already exists in our system
		if p, err := data.DB.Product.FindOne(db.Cond{
			"place_id":    place.ID,
			"external_id": p.ProductID,
		}); err != nil && err != db.ErrNoMoreRows {
			return errors.Wrap(err, "failed to fetch product")
		} else if p != nil {
			product.ID = p.ID
		}
		//if whType == shopify.TopicProductListingsUpdate && product.ID == 0 {
		//lg.Alertf("product list update with unknown external id: %d. Did not update.", p.ProductID)
		//return nil
		//}

		// parse product images
		for _, img := range p.Images {
			imgUrl, _ := url.Parse(img.Src)
			imgUrl.Scheme = "https"
			if img.Position == 1 {
				product.ImageUrl = imgUrl.String()
			}
			// always add to etc
			product.Etc.Images = append(product.Etc.Images, imgUrl.String())
		}

		variants := make([]*data.ProductVariant, len(p.Variants))
		var variantPrice float64
		for i, v := range p.Variants {
			price, _ := strconv.ParseFloat(v.Price, 64)
			variantPrice += price
			etc := data.ProductVariantEtc{
				Price: price,
				Sku:   v.Sku,
			}
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
			variants[i] = &data.ProductVariant{
				PlaceID:     place.ID,
				ProductID:   product.ID,
				OfferID:     v.ID,
				Description: v.Title,
				Limits:      int64(v.InventoryQuantity),
				Etc:         etc,
			}
			// fetch variant via offerID
			if vv, _ := data.DB.ProductVariant.FindOne(db.Cond{"offer_id": v.ID}); vv != nil {
				variants[i].ID = vv.ID
			}
		}
		// average variant price
		variantPrice = variantPrice / float64(len(p.Variants))

		if err := data.DB.Product.Save(product); err != nil {
			return errors.Wrap(err, "failed to save product")
		}

		for _, v := range variants {
			v.ProductID = product.ID
			if err := data.DB.ProductVariant.Save(v); err != nil {
				lg.Warn(errors.Wrap(err, "failed to save product variants"))
				continue
			}
		}

		tags := parseTags(p.Tags)
		q := data.DB.InsertInto("product_tags").
			Columns("product_id", "place_id", "value", "type").
			Amend(func(query string) string {
				return query + ` ON CONFLICT DO NOTHING`
			})
		b := q.Batch(len(tags))
		go func() {
			defer b.Done()

			foundGender := false
			for _, t := range tags {
				// detect if one of "man" or "woman"
				// TODO: let's do this some other way
				// and detect more product tag types
				typ := data.ProductTagTypeGeneral
				if t == "man" || t == "woman" || t == "unisex" {
					typ = data.ProductTagTypeGender
					foundGender = true
				}
				b.Values(product.ID, place.ID, t, typ)
			}

			// Product type category or gender
			for _, t := range parseTags(p.ProductType) {
				typ := data.ProductTagTypeCategory
				if t == "man" || t == "woman" || t == "unisex" {
					// skip gender if specified, if we've already found gender
					if foundGender {
						continue
					}
					typ = data.ProductTagTypeGender
					foundGender = true
				}
				b.Values(product.ID, place.ID, t, typ)
			}

			// Pull and parse collections, for now just parse gender
			// TODO: categories.
			//if !foundGender {
			//clist, _ := getProductCollections(ctx, p.ProductID)
			//for _, c := range clist {
			//for _, ctag := range parseTags(c.Handle) {
			//if ctag == "man" || ctag == "woman" || ctag == "unisex" {
			//foundGender = true
			//b.Values(product.ID, place.ID, ctag, data.ProductTagTypeGender)
			//}
			//}
			//}
			//}

			// Product Vendor/Brand
			b.Values(product.ID, place.ID, p.Vendor, data.ProductTagTypeBrand)

			// Average variant prices
			b.Values(product.ID, place.ID, fmt.Sprintf("%.2f", variantPrice), data.ProductTagTypePrice)

			// Variant options (ie. Color, Size, Material)
			for _, o := range p.Options {
				var typ data.ProductTagType
				typ.UnmarshalText([]byte(strings.ToLower(o.Name)))

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
		}()
		if err := b.Wait(); err != nil {
			lg.Warn(err)
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

var tagRegex = regexp.MustCompile("[^a-zA-Z0-9-]+")

func parseTags(tagStr string, optTags ...string) []string {
	tt := tagRegex.Split(tagStr, -1)

	tagSet := set.New()
	for _, t := range tt {
		t = strings.ToLower(t)

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
		tagSet.Add(strings.ToLower(t))
	}

	return set.StringSlice(tagSet)
}
