package data

import (
	"time"

	"upper.io/bond"
	"upper.io/db.v2"
)

type Like struct {
	ID        int64      `db:"id,pk,omitempty" json:"id"`
	UserID    int64      `db:"user_id" json:"userId"`
	PostID    int64      `db:"post_id" json:"postId"`
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

func (store *LikeStore) FindByPostID(postID int64) ([]*Like, error) {
	return store.FindAll(db.Cond{"post_id": postID})
}

func (store *LikeStore) FindByID(likeID int64) (*Like, error) {
	return store.FindOne(db.Cond{"id": likeID})
}

func (store *LikeStore) FindAll(cond db.Cond) ([]*Like, error) {
	var ls []*Like
	if err := store.Find(cond).All(&ls); err != nil {
		return nil, err
	}
	return ls, nil
}

func (store *LikeStore) FindOne(cond db.Cond) (*Like, error) {
	var l *Like
	if err := store.Find(cond).One(&l); err != nil {
		return nil, err
	}
	return l, nil
}

func (l *Like) BeforeCreate(bond.Session) error {
	l.CreatedAt = GetTimeUTCPointer()
	return nil
}

func (l *Like) AfterCreate(sess bond.Session) error {
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
