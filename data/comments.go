package data

import (
	"time"

	"upper.io/bond"
	"upper.io/db.v2"
)

type Comment struct {
	ID        int64      `db:"id,pk,omitempty" json:"id"`
	UserID    int64      `db:"user_id" json:"userId"`
	PostID    int64      `db:"post_id" json:"postId"`
	Body      string     `db:"body" json:"body"`
	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt,omitempty"`
}

type CommentStore struct {
	bond.Store
}

var _ interface {
	bond.HasBeforeCreate
	bond.HasAfterCreate
	bond.HasAfterDelete
} = &Comment{}

func (c *Comment) CollectionName() string {
	return `comments`
}

func (store *CommentStore) FindByPostID(postID int64) ([]*Comment, error) {
	return store.FindAll(db.Cond{"post_id": postID})
}

func (store *CommentStore) FindByID(commentID int64) (*Comment, error) {
	return store.FindOne(db.Cond{"id": commentID})
}

func (store *CommentStore) FindAll(cond db.Cond) ([]*Comment, error) {
	var cs []*Comment
	if err := store.Find(cond).All(&cs); err != nil {
		return nil, err
	}
	return cs, nil
}

func (store *CommentStore) FindOne(cond db.Cond) (*Comment, error) {
	var c *Comment
	if err := store.Find(cond).One(&c); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Comment) BeforeCreate(bond.Session) error {
	c.CreatedAt = GetTimeUTCPointer()
	return nil
}

func (c *Comment) AfterCreate(bond.Session) error {
	go func() {
		if post, err := DB.Post.FindByID(c.PostID); err == nil {
			post.UpdateCommentCount()
		}
	}()
	return nil
}

func (c *Comment) AfterDelete(bond.Session) error {
	go func() {
		if post, err := DB.Post.FindByID(c.PostID); err == nil {
			post.UpdateCommentCount()
		}
	}()
	return nil
}
