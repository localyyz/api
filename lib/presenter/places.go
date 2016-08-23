package presenter

import (
	"context"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/pkg/errors"
	"upper.io/db.v2"
)

type Place struct {
	*data.Place
	Locale string `json:"locale"`
}

type PlaceWithPost struct {
	*Place
	Posts []*Post `json:"posts"`
}

type PlaceWithPromo struct {
	*Place
	Promo *data.Promo `json:"promo"`
}

func NewPlace(ctx context.Context, place *data.Place) (*Place, error) {
	locale, err := data.DB.Locale.FindByID(place.LocaleID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to presente place(%v) locale", place.ID)
	}
	return &Place{
		Place:  place,
		Locale: locale.Name,
	}, nil
}

func PlacesWithPromos(ctx context.Context, places ...*data.Place) ([]*PlaceWithPromo, error) {
	presented := make([]*PlaceWithPromo, len(places))
	for i, p := range places {
		promo, err := data.DB.Promo.FindByPlaceID(p.ID)
		if err != nil && err != db.ErrNoMoreRows {
			return nil, errors.Wrapf(err, "failed to present place(%v) promo", p.ID)
		}
		presented[i] = &PlaceWithPromo{Place: &Place{Place: p}, Promo: promo}
	}
	return presented, nil
}

func PlacesWithPosts(ctx context.Context, places ...*data.Place) ([]*PlaceWithPost, error) {
	presented := make([]*PlaceWithPost, len(places))

	for i, pl := range places {
		var posts []*data.Post
		err := data.DB.Post.
			Find(db.Cond{"place_id": pl.ID}).
			OrderBy("-score").
			Limit(5).
			All(&posts)
		if err != nil {
			return presented, errors.Wrapf(err, "failed to present place(%v) posts", pl.ID)
		}
		postPresented, err := Posts(ctx, posts...)
		if err != nil {
			return presented, err
		}
		place, err := NewPlace(ctx, pl)
		if err != nil {
			return presented, err
		}
		presented[i] = &PlaceWithPost{
			Place: place,
			Posts: postPresented,
		}
	}

	return presented, nil
}
