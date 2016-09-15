package presenter

import (
	"context"

	"github.com/goware/lg"
	"github.com/pkg/errors"
	"upper.io/db.v2"

	"bitbucket.org/moodie-app/moodie-api/data"
)

type Post struct {
	*data.Post
	User    *data.User   `json:"user"`
	Place   *data.Place  `json:"place"`
	Promo   *data.Promo  `json:"promo"`
	Context *UserContext `json:"context"`
}

func NewPost(post *data.Post) *Post {
	return &Post{Post: post}
}

func (p *Post) WithUser() *Post {
	var err error
	if p.User, err = data.DB.User.FindByID(p.UserID); err != nil {
		lg.Error(errors.Wrapf(err, "failed to present post(%v) user", p.ID))
	}
	return p
}

func Posts(ctx context.Context, posts ...*data.Post) ([]*Post, error) {
	presented := make([]*Post, len(posts))
	for i, p := range posts { // TODO: bulk
		user, err := data.DB.User.FindByID(p.UserID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to present post(%v) user", p.ID)
		}
		liked, err := data.DB.Like.Find(db.Cond{"user_id": user.ID, "post_id": p.ID}).Count()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to present post(%v) like", p.ID)
		}
		presented[i] = &Post{
			Post: p,
			User: user,
			Context: &UserContext{
				Liked: (liked > 0),
			},
		}
	}
	return presented, nil
}

func PostsWithPlaces(ctx context.Context, posts ...*data.Post) ([]*Post, error) {
	presented, err := Posts(ctx, posts...)
	if err != nil {
		return nil, err
	}

	for _, p := range presented {
		place, err := data.DB.Place.FindByID(p.PlaceID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to present post(%v) place", p.ID)
		}
		p.Place = place
	}
	return presented, nil
}
