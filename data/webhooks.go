package data

import (
	"time"

	"upper.io/bond"
	db "upper.io/db.v3"
)

type Webhook struct {
	ID           int64      `db:"id,pk,omitempty"`
	PlaceID      int64      `db:"place_id"`
	Topic        string     `db:"topic"`
	ExternalID   int64      `db:"external_id"`
	LastSyncedAt *time.Time `db:"last_synced_at"`
	CreatedAt    *time.Time `db:"created_at,omitempty"`
}

type WebhookStore struct {
	bond.Store
}

func (wh *Webhook) CollectionName() string {
	return `webhooks`
}

func (store WebhookStore) FindByPlaceID(placeID int64) ([]*Webhook, error) {
	return store.FindAll(db.Cond{"place_id": placeID})
}

func (store WebhookStore) FindAll(cond db.Cond) ([]*Webhook, error) {
	var webhooks []*Webhook
	if err := store.Find(cond).All(&webhooks); err != nil {
		return nil, err
	}
	return webhooks, nil
}
