package data

import (
	"time"

	"upper.io/bond"
	"upper.io/db.v2"
)

// UserPoint keeps track of points awarded or used by the user
type UserPoint struct {
	ID     int64 `db:"id,pk,omitempty" json:"id"`
	UserID int64 `db:"user_id" json:"userId"`

	// Point could have been earned through posting a picture
	//    to a venue or earned through user engadgement
	// Point can be used by a user to peek at a promotion
	PostID  *int64 `db:"post_id,omitempty" json:"postId,omitempty"`
	PlaceID int64  `db:"place_id" json:"placeId"`
	PromoID int64  `db:"promo_id" json:"promoId"`
	PeekID  *int64 `db:"peek_id,omitempty" json:"peekId,omitempty"`

	Reward int64 `db:"reward" json:"reward"`

	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt,omitempty"`
	UpdatedAt *time.Time `db:"updated_at,omitempty" json:"updatedAt,omitempty"`
}

type UserPointStore struct {
	bond.Store
}

type UserPointPresenter struct {
	*UserPoint
	Place *Place `json:"place"`
	Post  *Post  `json:"post"`
	Promo *Promo `json:"promo"`
}

func (p *UserPoint) CollectionName() string {
	return `user_points`
}

func (store UserPointStore) CountByUserID(userID int64) (uint64, error) {
	return store.Find(db.Cond{"user_id": userID}).Count()
}

func (store UserPointStore) FindByUserID(userID int64) ([]*UserPoint, error) {
	return store.FindAll(db.Cond{"user_id": userID})
}

func (store UserPointStore) FindAll(cond db.Cond) ([]*UserPoint, error) {
	var points []*UserPoint
	if err := DB.UserPoint.Find(cond).OrderBy("-created_at").All(&points); err != nil {
		return nil, err
	}
	return points, nil
}
