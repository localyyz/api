package presenter

import (
	"github.com/goware/lg"
	"github.com/pkg/errors"

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
