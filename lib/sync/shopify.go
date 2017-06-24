package sync

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	set "gopkg.in/fatih/set.v0"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"github.com/gedex/inflector"
	"github.com/goware/lg"
	"github.com/pkg/errors"
)

func ShopifyProducts(ctx context.Context) error {
	list := ctx.Value("sync.list").([]*shopify.Product)
	place := ctx.Value("sync.place").(*data.Place)

	for _, p := range list {
		imgUrl, _ := url.Parse(p.Image.Src)
		imgUrl.Scheme = "https"

		product := &data.Product{
			PlaceID:     place.ID,
			ExternalID:  p.Handle,
			Title:       p.Title,
			Description: p.BodyHTML,
			ImageUrl:    imgUrl.String(),
		}
		var promos []*data.Promo

		for _, v := range p.Variants {
			price, _ := strconv.ParseFloat(v.Price, 64)
			promo := &data.Promo{
				PlaceID:     place.ID,
				ProductID:   product.ID,
				OfferID:     v.ID,
				Status:      data.PromoStatusActive,
				Description: v.Title,
				UserID:      0, // admin
				Limits:      int64(v.InventoryQuantity),
				Etc: data.PromoEtc{
					Price: price,
					Sku:   v.Sku,
				},
			}
			promos = append(promos, promo)
		}

		if err := data.DB.Product.Save(product); err != nil {
			return errors.Wrap(err, "failed to save promotion")
		}

		for _, v := range promos {
			v.ProductID = product.ID
			if err := data.DB.Promo.Save(v); err != nil {
				lg.Warn(errors.Wrap(err, "failed to save promotion"))
				continue
			}
		}

		tags := parseTags(p.Tags, p.ProductType, p.Vendor)
		q := data.DB.InsertInto("product_tags").Columns("product_id", "value", "type")
		b := q.Batch(len(tags))
		go func() {
			defer b.Done()
			// General product tags
			for _, t := range tags {
				// detect if one of "man" or "woman"
				// TODO: let's do this some other way
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

				for _, v := range o.Values {
					b.Values(product.ID, strings.ToLower(v), typ)
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
