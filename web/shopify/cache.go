package shopify

import (
	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
)

var storeCache map[string]*data.Place

func SetupShopCache(places ...*data.Place) {
	storeCache = make(map[string]*data.Place)
	for _, p := range places {
		storeCache[p.ShopifyID] = p
	}
	lg.Printf("shopify wh cache: keys(%d)", len(storeCache))
}

func cacheGet(key string) (*data.Place, error) {
	place, ok := storeCache[key]
	if !ok {
		var err error
		place, err = data.DB.Place.FindOne(
			db.Cond{
				"shopify_id": key,
				"status":     data.PlaceStatusActive,
			},
		)
		if err != nil {
			return nil, err
		}
		storeCache[place.ShopifyID] = place
	}
	return place, nil
}
