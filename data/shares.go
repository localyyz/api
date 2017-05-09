package data

import (
	"time"

	"upper.io/bond"
)

type Share struct {
	ID      int64 `db:"id,pk,omitempty" json:"id,omitempty"`
	UserID  int64 `db:"user_id" json:"userId"`
	PlaceID int64 `db:"place_id" json:"placeId"`

	Network        string `db:"network" json:"network"`
	NetworkShareID string `db:"network_share_id" json:"-"`
	Reach          int32  `db:"reach" json:"reach"`

	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt,omitempty"`
}

var _ interface {
	bond.HasBeforeCreate
} = &Share{}

type ShareStore struct {
	bond.Store
}

func (s *Share) CollectionName() string {
	return `shares`
}

func (s *Share) BeforeCreate(bond.Session) error {
	s.CreatedAt = GetTimeUTCPointer()
	return nil
}
