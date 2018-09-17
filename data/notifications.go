package data

import (
	"time"

	"upper.io/bond"
	db "upper.io/db.v3"
)

type Notification struct {
	ID          int64      `db:"id,pk,omitempty" json:"id"`
	CartID      *int64     `db:"cart_id,omitempty"`
	UserID      int64      `db:"user_id"`
	ProductID   int64      `db:"product_id"`
	VariantID   *int64     `db:"variant_id,omitempty"`
	ExternalID  string     `db:"external_id"`
	Heading     string     `db:"heading"`
	Content     string     `db:"content"`
	ScheduledAt *time.Time `db:"scheduled_at,omitempty"`
}

type NotificationStore struct {
	bond.Store
}

func (c *Notification) CollectionName() string {
	return `notifications`
}

func (store NotificationStore) FindOne(cond db.Cond) (*Notification, error) {
	var notf *Notification
	if err := DB.Notification.Find(cond).One(&notf); err != nil {
		return nil, err
	}
	return notf, nil
}

func (store NotificationStore) FindAll(cond db.Cond) ([]*Notification, error) {
	var notfs []*Notification
	if err := DB.Notification.Find(cond).All(&notfs); err != nil {
		return nil, err
	}
	return notfs, nil
}
