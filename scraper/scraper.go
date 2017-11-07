package scraper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/htmlx"
	"github.com/goware/geotools"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
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

type scrape struct {
	url  string
	ID   string
	name string
}

var scrapes = []scrape{
	{"shoplostfound.com/collections/new-arrivals", "lostandfound", "Lost&Found"},
	{"bookhou.com/collections/tote", "shopbookhou", "Bookhou"},
	{"theseptember.com/collections/just-in", "the-september", "The September"},
	{"gussiedup.ca/collections/new-arrivals", "gussied-up-plus", "Gussied up"},
	{"narwhalboutique.com/collections/whats-new/what&#39;s-new", "thenarwhalboutique", "The Narwhal Boutique"},
	{"treccanimilano.com/collections/luxury-leather-riding-style-boots-buy-online", "treccanimilano", "Treccani Milano"},
	{"comrags.com/collections/new-arrivals-1", "comrags", "Comrags"},
	{"modernvice.com/collections/fall-collection-2017", "modernvice", "Modern Vice"},
	{"botticellishoes.com/collections/women-sale", "botticellishoes", "Botticelli"},
	{"botkier.com/collections/new-arrivals", "botkier", "Botkier"},
	{"rue107.com/collections/the-new", "rue107-shop", "Rue 107"},
	{"mystiqueboutiquenyc.com/collections/new-arrivals", "mystiqueboutiquenyc", "Mystique Boutique NYC"},
	{"necessaryclothing.com/collections/new-in", "necessary-clothing-store", "Necessary Clothing"},
	{"ayr.com/collections/everything", "ayr-production", "AYR"},
	{"publicschoolnyc.com/collections/women-new-arrivals", "public-school", "Public School NYC"},
	{"thefrankieshop.com/collections/new-arrivals", "the-frankie-shop", "The Frankie Shop"},
	{"reasonclothing.com/collections/women/New-Arrivals", "reason-clothing", "Reason Clothing"},
	{"sundaysomewhere.com/collections/new-in", "sundaysomewhere", "Sunday Somewhere"},
	{"odinnewyork.com/collections/acne-studios", "odin-new-york", "Odin NYC"},
}

func ShopifyScraper() {
	for _, s := range scrapes {
		place, err := data.DB.Place.FindOne(db.Cond{"shopify_id": s.ID})
		if err != nil {
			if err != db.ErrNoMoreRows {
				continue
			}
			place = &data.Place{
				Name:      s.name,
				LocaleID:  0,
				ShopifyID: s.ID,
				Status:    data.PlaceStatusActive,
				Geo:       *geotools.NewPoint(0, 0),
			}
			if err := data.DB.Place.Save(place); err != nil {
				lg.Alertf("scraper: place %s failed with %+v", s.name, err)
				continue
			}
		}

		resp, err := http.Get(fmt.Sprintf("https://%s.oembed", s.url))
		if err != nil {
			lg.Warn(err)
			continue
		}

		var wrapper struct {
			Products []*Product `json:"products"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
			continue
		}

		for _, p := range wrapper.Products {
			// check if product already exists
			product, err := data.DB.Product.FindOne(db.Cond{
				"place_id":    place.ID,
				"external_id": p.ProductID,
			})

			if err != nil {
				if err != db.ErrNoMoreRows {
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
					Description: strings.TrimSpace(htmlx.StripTags(p.Description)),
					ImageUrl:    u.String(),
					Etc: data.ProductEtc{
						Brand:  strings.ToLower(p.Brand),
						Images: []string{u.String()},
					},
				}
				if err := data.DB.Product.Save(product); err != nil {
					continue
				}
			}

			// update offers/variants
			for _, o := range p.Offers {
				v, err := data.DB.ProductVariant.FindByOfferID(o.OfferID)
				if err != nil && err != db.ErrNoMoreRows {
					continue
				}

				if v != nil && o.InStock {
					// skip if inStock has not changed
					continue
				}

				if v == nil {
					// new variant
					v = &data.ProductVariant{
						PlaceID:     place.ID,
						OfferID:     o.OfferID,
						ProductID:   product.ID,
						Description: strings.ToLower(o.Title),
						Limits:      1,
						Etc: data.ProductVariantEtc{
							Price: o.Price,
							Sku:   o.Sku,
						},
					}
				}
				if !o.InStock {
					v.Limits = 0
				}

				data.DB.ProductVariant.Save(v)
			}
		}

		resp.Body.Close()
	}
}
