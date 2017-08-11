package sync

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	set "gopkg.in/fatih/set.v0"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/htmlx"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"github.com/gedex/inflector"
	"github.com/goware/lg"
	"github.com/pkg/errors"
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
			Etc:         data.ProductEtc{},
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
		for i, v := range p.Variants {
			price, _ := strconv.ParseFloat(v.Price, 64)
			etc := data.ProductVariantEtc{
				Price: price,
				Sku:   v.Sku,
			}
			// variant option values
			for _, o := range v.OptionValues {
				v := strings.ToLower(o.Value)
				switch o.Name {
				case "Size":
					etc.Size = v
				case "Color":
					etc.Color = v
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

		}

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

		tags := parseTags(p.Tags, p.ProductType, p.Vendor)
		q := data.DB.InsertInto("product_tags").Columns("product_id", "value", "type")
		b := q.Batch(len(tags))
		go func() {
			defer b.Done()
			for _, t := range tags {
				// detect if one of "man" or "woman"
				// TODO: let's do this some other way
				// and detect more product tag types
				typ := data.ProductTagTypeGeneral
				if t == "man" || t == "woman" {
					typ = data.ProductTagTypeGender
				}
				b.Values(product.ID, t, typ)
			}

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
					b.Values(product.ID, vv, typ)
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

func parseTags(tagStr string, optTags ...string) []string {
	tt := strings.FieldsFunc(tagStr, tagSplit)
	tagSet := set.New()
	for _, t := range tt {
		t = strings.TrimSpace(t)
		t = strings.ToLower(t)

		tt := inflector.Singularize(t)
		for {
			if tt == t {
				break
			}
			t = tt
			tt = inflector.Singularize(t)
		}
		tagSet.Add(t)
	}
	for _, t := range optTags {
		tagSet.Add(strings.ToLower(t))
	}

	return set.StringSlice(tagSet)
}

func tagSplit(r rune) bool {
	return r == ',' || r == ' '
}
