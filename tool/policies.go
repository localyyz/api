package tool

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	db "upper.io/db.v3"
)

func GetPolicies(w http.ResponseWriter, r *http.Request) {
	// producer
	var places []*data.Place
	err := data.DB.Place.Find(
		db.Cond{
			"status": data.PlaceStatusActive,
		},
	).OrderBy("-id").All(&places)
	if err != nil {
		log.Println(err)
		return
	}

	ctx := r.Context()
	for _, p := range places {
		if err := getShippingPolicy(ctx, p); err != nil {
			log.Println(err)
		}
		if err := getReturnPolicy(p); err != nil {
			log.Println(err)
		}
	}
}

var allowedCountrys = []string{"US", "CA"}

func hasAllowCountry(countries []*shopify.ShippingZoneCountry) bool {
	for _, c := range countries {
		for _, a := range allowedCountrys {
			if c.Code == a {
				return true
			}
		}
	}
	return false
}

func getShippingPolicy(ctx context.Context, p *data.Place) error {
	client, err := connect.GetShopifyClient(p.ID)
	if err != nil {
		return err
	}
	zones, _, err := client.ShippingZone.List(ctx)
	if err != nil {
		return err
	}
	log.Printf("found %d zones for %s", len(zones), p.Name)
	for _, z := range zones {
		// check if this zone is US / Canada.
		if !hasAllowCountry(z.Countries) {
			log.Printf("skipped zone %s", z.Name)
			continue
		}

		for _, c := range z.Countries {
			if c.Code != "US" && c.Code != "CA" {
				log.Printf("skipped country %s", c.Name)
				continue
			}
			sz := data.ShippingZone{
				PlaceID:    p.ID,
				Name:       z.Name,
				ExternalID: z.ID,
				Country:    strings.ToLower(c.Name),
			}
			for _, r := range c.Provinces {
				sz.Regions = append(sz.Regions, data.Region{
					Region:     r.Name,
					RegionCode: r.Code,
				})
			}

			for _, wz := range z.WeightBasedShippingRates {
				wpz := sz
				wpz.Type = data.ShippingZoneTypeByWeight
				wpz.Description = wz.Name
				wpz.WeightLow = wz.WeightLow
				wpz.WeightHigh = wz.WeightHigh
				wpz.Price, _ = strconv.ParseFloat(wz.Price, 64)
				data.DB.ShippingZone.Save(&wpz)
			}

			for _, pz := range z.PriceBasedShippingRates {
				ppz := sz
				ppz.Type = data.ShippingZoneTypeByPrice
				ppz.Description = pz.Name
				ppz.SubtotalLow, _ = strconv.ParseFloat(pz.MinOrderSubtotal, 64)
				ppz.SubtotalHigh, _ = strconv.ParseFloat(pz.MaxOrderSubtotal, 64)
				ppz.Price, _ = strconv.ParseFloat(pz.Price, 64)
				data.DB.ShippingZone.Save(&ppz)
			}
		}

	}

	return nil
}

func getReturnPolicy(p *data.Place) error {
	client, _ := connect.GetShopifyClient(p.ID)
	policies, _, err := client.Policy.List(context.Background())
	if err != nil {
		log.Printf("policy fetch error for place(%d, %s): %+v", p.ID, p.Name, err)
		return err
	}
	for _, po := range policies {
		if po.Title == shopify.PolicyRefund {
			p.ReturnPolicy.Description = po.Body
			if p.ReturnPolicy.URL == "" {
				p.ReturnPolicy.URL = po.URL
			}
			data.DB.Place.Save(p)
			log.Printf("saved refund policy for %s", p.Name)
		}
	}

	return nil
}
