package presenter

import (
	"context"

	"github.com/pkg/errors"

	db "upper.io/db.v2"

	"bitbucket.org/moodie-app/moodie-api/data"
)

type Promo struct {
	*data.Promo
	Post    *Post        `json:"post"`
	Context *UserContext `json:"context"`
}

func NewPromo(ctx context.Context, promo *data.Promo) (*Promo, error) {
	user := ctx.Value("session.user").(*data.User)

	presented := &Promo{Promo: promo}
	count, err := data.DB.Post.Find(db.Cond{"promo_id": promo.ID, "user_id": user.ID}).Count()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to present promo(%v) user context", promo.ID)
	}
	presented.Context = &UserContext{Promoted: (count > 0)}

	var post *data.Post
	err = data.DB.Post.
		Find(db.Cond{"promo_id": promo.ID, "promo_status": data.RewardCompleted}).
		OrderBy("-created_at").
		One(&post)
	if err != nil {
		if err == db.ErrNoMoreRows {
			return presented, nil
		}
		return nil, errors.Wrapf(err, "failed to present promo(%v) post", promo.ID)
	}
	presented.Post, err = NewPost(ctx, post)
	if err != nil {
		return nil, err
	}

	return presented, nil
}
