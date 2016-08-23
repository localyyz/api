package presenter

import (
	"context"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/pkg/errors"
	"upper.io/db.v2"
)

type PlaceWithPost struct {
	*data.Place
	Posts    []*Post `json:"posts"`
	Distance float64 `json:"distance"`
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
		presented[i] = &PlaceWithPost{
			Place: pl,
			Posts: postPresented,
		}
	}

	return presented, nil
}
