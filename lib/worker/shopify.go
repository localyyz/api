package worker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/goware/lg"
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

func TrackShopifySales() {
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

		resp, err := http.Get(fmt.Sprintf("%s.oembed", t.SalesUrl))
		if err != nil {
			lg.Warn(err)
			continue
		}

		var wrapper struct {
			Products []*Product `json:"products"`
			Provider string     `json:"provider"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
			lg.Fatal(err)
		}

		type listedPromo struct {
			*data.Promo
			IsValid bool
		}
		// get list of existing promotions that's active
		var existingPromos []*listedPromo
		if err := data.DB.Promo.Find(db.And(
			db.Cond{"place_id": place.ID},
			db.Or(
				db.Cond{"status": data.PromoStatusActive},
				db.Cond{"status": data.PromoStatusScheduled},
			),
		)).All(&existingPromos); err != nil {
			lg.Warn(errors.Wrap(err, "failed to fetch existing promotions"))
		}
		existingPromoMap := map[int64]*listedPromo{}
		for _, ep := range existingPromos {
			// mapped by offer id
			existingPromoMap[ep.OfferID] = ep
		}

		for _, p := range wrapper.Products {
			// check if product already exists
			product, err := data.DB.Product.FindOne(db.Cond{
				"place_id":    place.ID,
				"external_id": p.ProductID,
			})
			if err != nil {
				if err != db.ErrNoMoreRows {
					lg.Warn(err)
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
				}
				if err := data.DB.Product.Save(product); err != nil {
					lg.Warn(err)
					continue
				}
			}

			for _, o := range p.Offers {
				if promoWrapper, ok := existingPromoMap[o.OfferID]; ok {
					if o.InStock {
						// mark as still valid
						promoWrapper.IsValid = true
					}
					continue
				}

				if !o.InStock {
					continue
				}

				// new promotion
				now := time.Now().UTC()
				start := now.Add(5 * time.Minute)
				end := now.Add(30 * 24 * time.Hour)
				promo := &data.Promo{
					PlaceID:     place.ID,
					Type:        data.PromoTypePrice,
					OfferID:     o.OfferID,
					ProductID:   product.ID,
					Status:      data.PromoStatusActive,
					Description: o.Title,
					UserID:      0, // admin
					Etc: data.PromoEtc{
						Price: o.Price,
						Sku:   o.Sku,
					},
					StartAt: &start,
					EndAt:   &end, // 1 month
				}
				lg.Warnf("inserting %+v", promo)

				if err := data.DB.Promo.Save(promo); err != nil {
					lg.Warn(err)
					continue
				}
			}
		}

		var offerIDs []int64
		for _, p := range existingPromoMap {
			if !p.IsValid {
				offerIDs = append(offerIDs, p.OfferID)
			}
		}
		if len(offerIDs) > 0 {
			lg.Infof("Expiring offers: %+v", offerIDs)
			updateQuery := data.DB.Update("promos").
				Set(db.Cond{"status": data.PromoStatusCompleted}).
				Where(db.Cond{"offer_id": offerIDs})
			if _, err := updateQuery.Exec(); err != nil {
				lg.Warn(err)
			}
		}

		resp.Body.Close()
	}
	if err := q.Err(); err != nil {
		lg.Warn(err)
	}
	q.Close()
}
