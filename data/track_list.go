package data

import (
	"time"

	"upper.io/bond"
)

type TrackList struct {
	ID            int64      `db:"id"`
	PlaceID       int64      `db:"place_id"`
	SalesUrl      string     `db:"sales_url"`
	LastTrackedAt *time.Time `db:"last_tracked_at"`
	CreatedAt     *time.Time `db:"created_at"`
}

type TrackListStore struct {
	bond.Store
}

func (t *TrackList) CollectionName() string {
	return `track_list`
}
