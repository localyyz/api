package data

import (
	"time"

	"upper.io/bond"
)

type Comment struct {
	ID        int64      `db:"id,pk,omitempty" json:"id"`
	UserID    int64      `db:"user_id" json:"user_id"`
	PostID    int64      `db:"post_id" json:"post_id"`
	Body      string     `db:"body" json:"body"`
	CreatedAt *time.Time `db:"created_at,omitempty" json:"created_at,omitempty"`
}

type CommentStore struct {
	bond.Store
}

func (c *Comment) CollectionName() string {
	return `comments`
}
