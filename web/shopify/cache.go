package shopify

import (
	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
)

var storeCache map[string]*data.Place

func SetupShopCache(DB *data.Database) {
	places, err := DB.Place.FindAll(db.Cond{"status": data.PlaceStatusActive})
	if err != nil {
		lg.Alert("failed to cache place id at init")
		return
	}

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
		place, err = data.DB.Place.FindByShopifyID(key)
		if err != nil {
			return nil, err
		}
		storeCache[place.ShopifyID] = place
	}
	return place, nil
}
