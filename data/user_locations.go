package data

import (
	"time"

	"upper.io/bond"

	"github.com/goware/geotools"
)

type UserLocation struct {
	ID     int64          `db:"id,pk,omitempty" json:"id" facebook:"-"`
	UserID int64          `db:"user_id" json:"userId"`
	Geo    geotools.Point `db:"geo" json:"-"`

	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt,omitempty"`
}

type UserLocationStore struct {
	bond.Store
}

var _ interface {
	bond.HasBeforeCreate
} = &UserLocation{}

func (u *UserLocation) CollectionName() string {
	return `user_locations`
}

func (u *UserLocation) BeforeCreate(bond.Session) error {
	u.CreatedAt = GetTimeUTCPointer()
	return nil
}
