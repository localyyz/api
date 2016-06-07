package data

import (
	"time"

	"upper.io/bond"
	"upper.io/db"
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

var _ interface {
	bond.HasBeforeCreate
	bond.HasAfterCreate
	bond.HasAfterDelete
} = &Comment{}

func (c *Comment) CollectionName() string {
	return `comments`
}

func (s *CommentStore) FindByPostID(postID int64) ([]*Comment, error) {
	return s.FindAll(db.Cond{"post_id": postID})
}

func (s *CommentStore) FindByID(commentID int64) (*Comment, error) {
	return s.FindOne(db.Cond{"id": commentID})
}

func (s *CommentStore) FindAll(cond db.Cond) ([]*Comment, error) {
	var cs []*Comment
	if err := s.Find(cond).All(&cs); err != nil {
		return nil, err
	}
	return cs, nil
}

func (s *CommentStore) FindOne(cond db.Cond) (*Comment, error) {
	var c *Comment
	if err := s.Find(cond).One(&c); err != nil {
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
