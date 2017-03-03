package data

import "time"

type UserAccess struct {
	ID      int64 `db:"id,omitempty,pk" json:"ID"`
	UserID  int64 `db:"user_id,omitempty" json:"userId"`
	PlaceID int64 `db:"place_id,omitempty" json:"placeId"`

	// role definitions
	Admin    bool `db:"admin" json:"admin"`
	Promoter bool `db:"promoter" json:"promoter"`
	Member   bool `db:"member" json:"member"`

	CreatedAt *time.Time `db:"created_at,omitempty" json:"created_at"`
	UpdatedAt *time.Time `db:"updated_at,omitempty" json:"updated_at"`
}

type UserRoleType uint

func (ua *UserAccess) CollectionName() string {
	return `user_access`
}
