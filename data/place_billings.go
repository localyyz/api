package data

import (
	"fmt"
	"time"

	"upper.io/bond"
	db "upper.io/db.v3"
)

type PlaceBilling struct {
	ID         int64         `db:"id,pk,omitempty" json:"id,omitempty"`
	PlaceID    int64         `db:"place_id" json:"placeID"`
	PlanID     int64         `db:"plan_id" json:"planID"`
	ExternalID int64         `db:"external_id" json:"externalID"`
	Status     BillingStatus `db:"status" json:"status"`

	CreatedAt  *time.Time `db:"created_at,omitempty" json:"createdAt"`
	UpdatedAt  *time.Time `db:"updated_at,omitempty" json:"updatedAt"`
	AcceptedAt *time.Time `db:"accepted_at,omitempty" json:"acceptedAt"`
}

type BillingStatus uint32

type PlaceBillingStore struct {
	bond.Store
}

const (
	_                     BillingStatus = iota // 0
	BillingStatusPending                       // 1
	BillingStatusAccepted                      // 2
	// This is the only status that actually causes a merchant to be charged.
	// An accepted charge is transitioned to active via the activate endpoint.
	BillingStatusActive   // 3
	BillingStatusDeclined // 4
	// The recurring charge was not accepted within 2 days of being created.
	BillingStatusExpired // 5
	// The recurring charge is on hold due to a shop subscription non-payment.
	// The charge will re-activate once subscription payments resume.
	BillingStatusFrozen // 6
	// The developer cancelled the charge.
	BillingStatusCancelled // 7
)

var (
	billingStatuses = []string{
		"-",
		"pending",
		"accepted",
		"active",
		"declined",
		"expired",
		"frozen",
		"cancelled",
	}
)

var _ interface {
	bond.HasBeforeUpdate
} = &PlaceBilling{}

func (b *PlaceBilling) CollectionName() string {
	return `place_billings`
}

func (b *PlaceBilling) BeforeUpdate(bond.Session) error {
	if b.Status == BillingStatusActive {
		b.AcceptedAt = GetTimeUTCPointer()
	}
	return nil
}

func (store PlaceBillingStore) FindByPlaceID(placeID int64) (*PlaceBilling, error) {
	return store.FindOne(db.Cond{"place_id": placeID})
}

func (store PlaceBillingStore) FindOne(cond db.Cond) (*PlaceBilling, error) {
	var billing *PlaceBilling
	if err := store.Find(cond).One(&billing); err != nil {
		return nil, err
	}
	return billing, nil
}

// String returns the string value of the status.
func (s BillingStatus) String() string {
	return billingStatuses[s]
}

// MarshalText satisfies TextMarshaler
func (s BillingStatus) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalText satisfies TextUnmarshaler
func (s *BillingStatus) UnmarshalText(text []byte) error {
	enum := string(text)
	for i := 0; i < len(billingStatuses); i++ {
		if enum == billingStatuses[i] {
			*s = BillingStatus(i)
			return nil
		}
	}
	return fmt.Errorf("unknown billing status %s", enum)
}
