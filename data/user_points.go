package data

import (
	"time"

	"upper.io/bond"
	"upper.io/db"
)

// UserPoint keeps track of points awarded to the user
type UserPoint struct {
	ID     int64 `db:"id,pk,omitempty" json:"id"`
	UserID int64 `db:"user_id" json:"userId"`

	// Point could have been earned through posting a picture
	// to a venue or earned through user engadgement
	PostID  int64 `db:"post_id" json:"postId"`
	PlaceID int64 `db:"place_id" json:"placeId"`

	// internal multiplier associated with this point
	// multipliers are applied by promotions
	Multiplier uint32 `db:"multiplier" json:"-"`

	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt,omitempty"`
}

type UserPointStore struct {
	bond.Store
}

type UserPointPresenter struct {
	*UserPoint
	Place  *Place `json:"place"`
	Post   *Post  `json:"post"`
	Reward uint32 `json:"reward"`
}

var (
	DailyPointLimit = 3
)

func (p *UserPoint) CollectionName() string {
	return `user_points`
}

// TODO: smarter throttling. ie 2 points max per venue, up to 3 venue..
// IsLimited finds user points and checks if they have reached the daily point
// limit
func (store *UserPointStore) IsLimited(userID int64) (bool, error) {
	cond := db.Cond{
		"user_id":       userID,
		"created_at >=": time.Now().AddDate(0, 0, -1), // last 24
	}
	count, err := store.Find(cond).Count()
	if err != nil {
		return false, err
	}
	return (int(count) >= DailyPointLimit), nil
}

func (store *UserPointStore) FindByUserID(userID int64) ([]*UserPoint, error) {
	return store.FindAll(db.Cond{"user_id": userID})
}

func (store *UserPointStore) CountByUserID(userID int64) (uint64, error) {
	return store.Find(db.Cond{"user_id": userID}).Count()
}

func (store *UserPointStore) FindAll(cond db.Cond) ([]*UserPoint, error) {
	var points []*UserPoint
	if err := DB.UserPoint.Find(cond).Sort("-created_at").All(&points); err != nil {
		return nil, err
	}
	return points, nil
}
