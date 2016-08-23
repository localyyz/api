package presenter

import (
	"context"

	"github.com/pkg/errors"

	db "upper.io/db.v2"

	"bitbucket.org/moodie-app/moodie-api/data"
)

type Nearby struct {
	Places []*data.Place `json:"places"`
	Promos []*data.Promo `json:"promos"`
}

func NearbyPlaces(ctx context.Context, places ...*data.Place) (*Nearby, error) {
	// return any active promotions
	placeIDs := make([]int64, len(places))
	for i, p := range places {
		placeIDs[i] = p.ID
	}
	// query promos
	promos, err := data.DB.Promo.FindAll(db.Cond{"place_id": placeIDs})
	if err != nil {
		return nil, errors.Wrap(err, "failed to present nearby promo")
	}
	return &Nearby{Places: places, Promos: promos}, nil
}
