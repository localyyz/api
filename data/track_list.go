package data

import (
	"time"

	"upper.io/bond"
	db "upper.io/db.v3"
)

type TrackList struct {
	ID            int64      `db:"id"`
	PlaceID       int64      `db:"place_id"`
	SalesURL      string     `db:"sales_url"`
	LastTrackedAt *time.Time `db:"last_tracked_at"`
	CreatedAt     *time.Time `db:"created_at"`
}

type TrackListStore struct {
	bond.Store
}

func (t *TrackList) CollectionName() string {
	return `track_list`
}

func (store TrackListStore) FindByPlaceID(placeID int64) (*TrackList, error) {
	return store.FindOne(db.Cond{"place_id": placeID})
}

func (store TrackListStore) FindOne(cond db.Cond) (*TrackList, error) {
	var trackList *TrackList
	if err := store.Find(cond).One(&trackList); err != nil {
		return nil, err
	}
	return trackList, nil
}
