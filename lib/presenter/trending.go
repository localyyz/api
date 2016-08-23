package presenter

import (
	"context"

	"bitbucket.org/moodie-app/moodie-api/data"
)

type Trending struct {
	Nearby   []*PlaceWithPost `json:"nearby"`
	Promoted []*PlaceWithPost `json:"promoted"`
}

func TrendingPlaces(ctx context.Context, places ...*data.Place) (*Trending, error) {
	user := ctx.Value("session.user").(*data.User)

	presented := &Trending{[]*PlaceWithPost{}, []*PlaceWithPost{}}
	placePresented, err := PlacesWithPosts(ctx, places...)
	if err != nil {
		return presented, err
	}
	for _, pl := range placePresented {
		if pl.LocaleID == user.Etc.LocaleID {
			presented.Nearby = append(presented.Nearby, pl)
		} else {
			presented.Promoted = append(presented.Promoted, pl)
		}
	}
	return presented, nil
}
