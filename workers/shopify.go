package workers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/goware/lg"
	"github.com/mmcdole/gofeed"
	"github.com/pkg/errors"
)

type Offer struct {
	CurrencyCode string  `json:"currency_code"`
	InStock      bool    `json:"in_stock"`
	OfferID      int64   `json:"offer_id"`
	Price        float64 `json:"price"`
	Sku          string  `json:"sku"`
	Title        string  `json:"title"`
}

type Product struct {
	ProductID    string   `json:"product_id"`
	Brand        string   `json:"brand"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	ThumbnailURL string   `json:"thumbnail_url"`
	Offers       []*Offer `json:"offers"`
}

func getProductTags(salesURL string) map[string][]string {
	// we have to get the tags from atom feed.
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(fmt.Sprintf("%s.atom", salesURL))

	productTags := make(map[string][]string)
	for _, it := range feed.Items {
		pu, _ := url.Parse(it.Link)

		var tags []string
		for _, ext := range it.Extensions {
			for _, tag := range ext["tag"] {
				tags = append(tags, strings.ToLower(tag.Value))
			}
			if t, ok := ext["type"]; ok && len(t) > 0 {
				tags = append(tags, strings.ToLower(t[0].Value))
			}
			if t, ok := ext["vendor"]; ok && len(t) > 0 {
				tags = append(tags, strings.ToLower(t[0].Value))
			}
		}
		eID := strings.TrimPrefix(pu.Path, "/products/")
		productTags[eID] = tags
	}

	return productTags
}

func ShopifyPuller() {
	q := data.DB.TrackList.Find()
	for {
		var t *data.TrackList

		if !q.Next(&t) {
			break
		}

		t.LastTrackedAt = data.GetTimeUTCPointer()
		if err := data.DB.Save(t); err != nil {
			lg.Warn(err)
		}

		place, err := data.DB.Place.FindByID(t.PlaceID)
		if err != nil {
			lg.Warn(err)
			continue
		}

		resp, err := http.Get(fmt.Sprintf("%s.oembed", t.SalesURL))
		if err != nil {
			lg.Warn(err)
			continue
		}

		var wrapper struct {
			Products []*Product `json:"products"`
			Provider string     `json:"provider"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
			lg.Fatal(errors.Wrapf(err, "oembed parse error"))
		}

		// get product tags from atom feed
		productTags := getProductTags(t.SalesURL)

		for _, p := range wrapper.Products {
			// check if product already exists
			product, err := data.DB.Product.FindOne(db.Cond{
				"place_id":    place.ID,
				"external_id": p.ProductID,
			})

			if err != nil {
				if err != db.ErrNoMoreRows {
					lg.Warn(err)
					continue
				}

				u, err := url.Parse(p.ThumbnailURL)
				if err != nil {
					lg.Warn(err)
					continue
				}
				u.Scheme = "https"
				u.RawQuery = ""

				product = &data.Product{
					PlaceID:     place.ID,
					ExternalID:  p.ProductID,
					Title:       p.Title,
					Description: p.Description,
					ImageUrl:    u.String(),
					Etc: data.ProductEtc{
						Brand: strings.ToLower(p.Brand),
					},
				}
				if err := data.DB.Product.Save(product); err != nil {
					lg.Warn(err)
					continue
				}

				// batch insert tags
				tags, found := productTags[product.ExternalID]
				if !found {
					tags = []string{product.Etc.Brand}
				}

				q := data.DB.InsertInto("product_tags").Columns("product_id", "value")
				b := q.Batch(len(tags))
				go func() {
					defer b.Done()
					for _, t := range tags {
						b.Values(product.ID, t)
					}
				}()
				if err := b.Wait(); err != nil {
					lg.Warn(err)
				}
			}

			promos, err := data.DB.Product.FindPromos(product.ID)
			if err != nil {
				lg.Warn(errors.Wrapf(err, "prodocut(%d) promotions", product.ID))
				continue
			}
			dbPromoMap := make(map[int64]*data.Promo)
			for _, ep := range promos {
				// mapped by offer id
				ep.Status = data.PromoStatusCompleted
				dbPromoMap[ep.OfferID] = ep
			}

			// update promotions/offers/variants
			for _, o := range p.Offers {
				if p, ok := dbPromoMap[o.OfferID]; ok {
					if o.InStock {
						// mark as still valid
						p.Status = data.PromoStatusActive
					}
					continue
				}

				if !o.InStock {
					continue
				}

				// new promotion
				promo := &data.Promo{
					PlaceID:     place.ID,
					OfferID:     o.OfferID,
					ProductID:   product.ID,
					Status:      data.PromoStatusActive,
					Description: strings.ToLower(o.Title),
					Etc: data.PromoEtc{
						Price: o.Price,
						Sku:   o.Sku,
					},
				}
				lg.Warnf("inserting %+v", promo)

				if err := data.DB.Promo.Save(promo); err != nil {
					lg.Warn(err)
					continue
				}
			}

			// save expired promotions
			var expiredPromoIDs []int64
			for _, p := range dbPromoMap {
				if p.Status == data.PromoStatusCompleted {
					expiredPromoIDs = append(expiredPromoIDs, p.ID)
				}
			}
			if len(expiredPromoIDs) > 0 {
				updateQuery := data.DB.Update("promos").
					Set(db.Cond{"status": data.PromoStatusCompleted}).
					Where(db.Cond{"offer_id": expiredPromoIDs})
				if _, err := updateQuery.Exec(); err != nil {
					lg.Warn(err)
				}
			}
		}

		resp.Body.Close()
	}
	if err := q.Err(); err != nil {
		lg.Warn(err)
	}
	q.Close()
}
