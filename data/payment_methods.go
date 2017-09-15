package data

import (
	"time"

	"upper.io/bond"
	db "upper.io/db.v3"
)

type PaymentMethod struct {
	ID     int64 `db:"id" json:"id"`
	UserID int64 `db:"user_id" json:"userId"`

	Brand        string   `db:"brand" json:"brand"`
	ExpMonth     ExpMonth `db:"exp_month" json:"expMonth"`
	ExpYear      int32    `db:"exp_year" json:"expYear"`
	LastFour     string   `db:"last_four" json:"lastFour"`
	Country      string   `db:"country" json:"country"`
	StripeCardID string   `db:"stripe_card_id" json:"stripeCardId"`

	CreatedAt *time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt *time.Time `db:"updated_at" json:"updatedAt"`
	DeletedAt *time.Time `db:"deleted_at" json:"deletedAt"`
}

type ExpMonth uint32

type PaymentMethodStore struct {
	bond.Store
}

var _ interface {
	bond.HasBeforeCreate
} = &PaymentMethod{}

func (u *PaymentMethod) CollectionName() string {
	return `user_payment_methods`
}

func (u *PaymentMethod) BeforeCreate(bond.Session) error {
	u.CreatedAt = GetTimeUTCPointer()
	return nil
}

func (store PaymentMethodStore) FindByID(ID int64) (*PaymentMethod, error) {
	return store.FindOne(db.Cond{"id": ID})
}

func (store PaymentMethodStore) FindAllByUserID(userID int64) ([]*PaymentMethod, error) {
	return store.FindAll(db.Cond{"user_id": userID})
}

func (store PaymentMethodStore) FindOne(cond db.Cond) (*PaymentMethod, error) {
	var method *PaymentMethod
	if err := store.Find(cond).One(&method); err != nil {
		return nil, err
	}
	return method, nil
}

func (store PaymentMethodStore) FindAll(cond db.Cond) ([]*PaymentMethod, error) {
	var methods []*PaymentMethod
	if err := store.Find(cond).All(&methods); err != nil {
		return nil, err
	}
	return methods, nil
}
