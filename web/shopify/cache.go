package shopify

import (
	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
)

var (
	storeCache    map[string]*data.Place
	categoryCache map[string]*data.Category
)

func SetupShopCache(places ...*data.Place) {
	storeCache = make(map[string]*data.Place)
	for _, p := range places {
		storeCache[p.ShopifyID] = p
	}
	lg.Infof("store cache: keys(%d)", len(storeCache))
}

func SetupCategoryCache(categories ...*data.Category) {
	categoryCache = make(map[string]*data.Category)
	for _, c := range categories {
		categoryCache[c.Value] = c
	}
	lg.Infof("category cache: keys(%d)", len(categoryCache))
}

func storeGet(key string) (*data.Place, error) {
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
