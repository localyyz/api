package data

import (
	"time"

	"upper.io/bond"
	"upper.io/db"
)

type Like struct {
	ID        int64      `db:"id,pk,omitempty" json:"id"`
	UserID    int64      `db:"user_id" json:"user_id"`
	PostID    int64      `db:"post_id" json:"post_id"`
	CreatedAt *time.Time `db:"created_at,omitempty" json:"created_at,omitempty"`
}

type LikeStore struct {
	bond.Store
}

var _ interface {
	bond.HasBeforeCreate
	bond.HasAfterCreate
	bond.HasAfterDelete
} = &Like{}

func (l *Like) CollectionName() string {
	return `likes`
}

func (s *LikeStore) FindByPostID(postID int64) ([]*Like, error) {
	return s.FindAll(db.Cond{"post_id": postID})
}

func (s *LikeStore) FindByID(likeID int64) (*Like, error) {
	return s.FindOne(db.Cond{"id": likeID})
}

func (s *LikeStore) FindAll(cond db.Cond) ([]*Like, error) {
	var ls []*Like
	if err := s.Find(cond).All(&ls); err != nil {
		return nil, err
	}
	return ls, nil
}

func (s *LikeStore) FindOne(cond db.Cond) (*Like, error) {
	var l *Like
	if err := s.Find(cond).One(&l); err != nil {
		return nil, err
	}
	return l, nil
}

func (l *Like) BeforeCreate(bond.Session) error {
	l.CreatedAt = GetTimeUTCPointer()
	return nil
}

func (l *Like) AfterCreate(bond.Session) error {
	go func() {
		if post, err := DB.Post.FindByID(l.PostID); err == nil {
			post.UpdateLikeCount()
		}
	}()
	return nil
}

func (l *Like) AfterDelete(bond.Session) error {
	go func() {
		if post, err := DB.Post.FindByID(l.PostID); err == nil {
			post.UpdateLikeCount()
		}
	}()
	return nil
}
