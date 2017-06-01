package data

import (
	"context"
	"time"

	"upper.io/bond"
	"upper.io/db.v3"
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

func (store FollowingStore) IsFollowing(ctx context.Context, placeID int64) bool {
	user := ctx.Value("session.user").(*User)

	count, _ := store.Find(
		db.Cond{"place_id": placeID, "user_id": user.ID},
	).Count()

	return (count > 0)
}

func (store FollowingStore) FindByUserID(userID int64) ([]*Following, error) {
	return store.FindAll(db.Cond{"user_id": userID})
}

func (store FollowingStore) FindOne(cond db.Cond) (*Following, error) {
	var following *Following
	if err := store.Find(cond).One(&following); err != nil {
		return nil, err
	}
	return following, nil
}

func (store FollowingStore) FindAll(cond db.Cond) ([]*Following, error) {
	var followings []*Following
	if err := store.Find(cond).All(&followings); err != nil {
		return nil, err
	}
	return followings, nil
}
