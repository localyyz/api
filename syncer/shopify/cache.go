package shopify

import (
	"errors"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
)

var (
	storeCache       map[string]*data.Place
	ErrPlaceInactive = errors.New("place inactive")
	ErrPlaceNotFound = errors.New("place not found")
)

func SetupShopCache(places ...*data.Place) {
	storeCache = make(map[string]*data.Place)
	for _, p := range places {
		storeCache[p.ShopifyID] = p
	}
	lg.Infof("store cache: keys(%d)", len(storeCache))
}

func storeGet(key string) (*data.Place, error) {
	place, ok := storeCache[key]
	if !ok {
		var err error
		place, err = data.DB.Place.FindOne(
			db.Cond{"shopify_id": key},
		)
		if err != nil {
			return nil, err
		}
		storeCache[place.ShopifyID] = place
	}

	if place != nil && place.Status != data.PlaceStatusActive {
		return nil, ErrPlaceInactive
	}

	if place == nil {
		return nil, ErrPlaceNotFound
	}

	return place, nil
}
