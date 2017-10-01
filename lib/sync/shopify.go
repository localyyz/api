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

func ShopifyProductListings(ctx context.Context) error {
	list := ctx.Value("sync.list").([]*shopify.ProductList)
	place := ctx.Value("sync.place").(*data.Place)

	for _, p := range list {
		product := &data.Product{
			PlaceID:     place.ID,
			ExternalID:  p.Handle,
			Title:       p.Title,
			Description: strings.TrimSpace(htmlx.StripTags(p.BodyHTML)),
			Etc: data.ProductEtc{
				Brand: p.Vendor,
			},
		}

		// check if product already exists in our system
		if p, _ := data.DB.Product.FindOne(db.Cond{"external_id": p.Handle}); p != nil {
			product.ID = p.ID
		}

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
				Status:      data.ProductVariantStatusActive,
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
			for _, t := range tags {
				// detect if one of "man" or "woman"
				// TODO: let's do this some other way
				// and detect more product tag types
				typ := data.ProductTagTypeGeneral
				if t == "man" || t == "woman" || t == "unisex" {
					typ = data.ProductTagTypeGender
				}
				b.Values(product.ID, place.ID, t, typ)
			}

			// Product type category
			for _, t := range parseTags(p.ProductType) {
				if t == "man" || t == "woman" || t == "unisex" {
					// skip gender if specified
					continue
				}
				b.Values(product.ID, place.ID, t, data.ProductTagTypeCategory)
			}

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
