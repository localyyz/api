package presenter

import (
	"context"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/goware/lg"
	"github.com/pkg/errors"
	"upper.io/db.v2"
)

type Place struct {
	*data.Place
	Locale *data.Locale `json:"locale"`
	Posts  []*Post      `json:"posts"`
	Promo  *data.Promo  `json:"promo"`

	SneakReward    int64 `json:"sneakReward"`
	PromoCompleted bool  `json:"promoCompleted"`

	ctx context.Context
}

func NewPlace(ctx context.Context, place *data.Place) *Place {
	return &Place{Place: place, ctx: ctx}
}

// presents given place with locale detail
func (pl *Place) WithLocale() *Place {
	var err error
	if pl.Locale, err = data.DB.Locale.FindByID(pl.LocaleID); err != nil {
		lg.Error(errors.Wrapf(err, "failed to present place(%v) post", pl.ID))
	}
	return pl
}

// present the place with promotion
func (pl *Place) WithPromo() *Place {
	promo, err := data.DB.Promo.FindByPlaceID(pl.ID)
	if err != nil {
		lg.Error(errors.Wrapf(err, "failed to present place(%v) promo", pl.ID))
		return pl
	}
	pl.SneakReward = promo.Reward

	// TODO: have the user taken a sneakpeek?

	if pl.Distance < data.PromoDistanceLimit {
		pl.Promo = promo

		user := pl.ctx.Value("session.user").(*data.User)
		count, err := data.DB.Post.Find(db.Cond{"promo_id": promo.ID, "user_id": user.ID}).Count()
		if err != nil {
			lg.Warn(errors.Wrapf(err, "failed to present promo(%v) user context", promo.ID))
			return pl
		}
		pl.PromoCompleted = (count > 0)
	}
	return pl
}

// presents the place with the post with highest score, and it's promotion
func (pl *Place) WithPost() *Place {
	var post *data.Post
	err := data.DB.Post.Find(db.Cond{"place_id": pl.ID}).OrderBy("-score").One(&post)
	if err != nil {
		lg.Error(errors.Wrapf(err, "failed to present place(%v) post", pl.ID))
		return pl
	}

	// posts with user
	pl.Posts = []*Post{NewPost(post).WithUser()}
	return pl
}
