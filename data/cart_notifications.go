package data

import (
	"time"

	"upper.io/bond"
	db "upper.io/db.v3"
)

type CartNotification struct {
	ID          int64      `db:"id,pk,omitempty" json:"id"`
	CartID      int64      `db:"cart_id"`
	ProductID   int64      `db:"product_id"`
	VariantID   int64      `db:"variant_id"`
	ExternalID  string     `db:"external_id"`
	Heading     string     `db:"heading"`
	Content     string     `db:"content"`
	ScheduledAt *time.Time `db:"scheduled_at,omitempty"`

	// not saved
	UserID int64
}

type CartNotificationStore struct {
	bond.Store
}

func (c *CartNotification) CollectionName() string {
	return `cart_notifications`
}

func (store CartNotificationStore) FindOne(cond db.Cond) (*CartNotification, error) {
	var cart *CartNotification
	if err := DB.CartNotification.Find(cond).One(&cart); err != nil {
		return nil, err
	}
	return cart, nil
}

func (store CartNotificationStore) FindAll(cond db.Cond) ([]*CartNotification, error) {
	var carts []*CartNotification
	if err := DB.CartNotification.Find(cond).All(&carts); err != nil {
		return nil, err
	}
	return carts, nil
}
