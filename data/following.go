package data

import (
	"time"

	db "upper.io/db.v2"

	"github.com/upper/bond"
)

type Following struct {
	ID        int64      `db:"id,pk,omitempty" json:"id,omitempty"`
	UserID    int64      `db:"user_id" json:"userId"`
	PlaceID   int64      `db:"place_id" json:"placeId"`
	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt"`
}

type FollowingStore struct {
	bond.Store
}

func (f *Following) CollectionName() string {
	return `followings`
}

func (store FollowingStore) FindByUserID(userID int64) ([]*Following, error) {
	return store.FindAll(db.Cond{"user_id": userID})
}

func (store FollowingStore) FindAll(cond db.Cond) ([]*Following, error) {
	var followings []*Following
	if err := store.Find(cond).All(&followings); err != nil {
		return nil, err
	}
	return followings, nil
}
