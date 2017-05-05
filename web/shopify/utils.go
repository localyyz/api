package shopify

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
)

func getProductPromo(ctx context.Context, p *shopify.Product) (*data.Product, []*data.Promo) {
	place := ctx.Value("place").(*data.Place)

	imgUrl, _ := url.Parse(p.Image.Src)
	imgUrl.Scheme = "https"

	product := &data.Product{
		PlaceID:     place.ID,
		ExternalID:  fmt.Sprintf("%d", p.Handle),
		Title:       p.Title,
		Description: p.BodyHTML,
		ImageUrl:    imgUrl.String(),
	}
	product.ParseTags(p.Tags, p.ProductType, p.Vendor)

	var promos []*data.Promo
	for _, v := range p.Variants {
		now := time.Now().UTC()
		start := now.Add(1 * time.Minute)
		end := now.Add(30 * 24 * time.Hour)
		price, _ := strconv.ParseFloat(v.Price, 64)
		promo := &data.Promo{
			PlaceID:     place.ID,
			Type:        data.PromoTypePrice,
			OfferID:     v.ID,
			Status:      data.PromoStatusActive,
			Description: v.Title,
			UserID:      0, // admin
			Limits:      int64(v.InventoryQuantity),
			Etc: data.PromoEtc{
				Price: price,
				Sku:   v.Sku,
			},
			StartAt: &start,
			EndAt:   &end, // 1 month
		}
		promos = append(promos, promo)
	}

	return product, promos
}
