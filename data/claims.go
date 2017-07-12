package data

import (
	"errors"
	"fmt"
	"time"

	"upper.io/bond"
)

type Claim struct {
	ID      int64 `db:"id,pk,omitempty" json:"id,omitempty"`
	PromoID int64 `db:"promo_id" json:"promoId"`
	PlaceID int64 `db:"place_id" json:"placeId"`
	UserID  int64 `db:"user_id" json:"userId"`

	//TODO: Hash   string      `db:"hash" json:"hash"`
	Status ClaimStatus `db:"status" json:"status"`

	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt"`
	UpdatedAt *time.Time `db:"updated_at,omitempty" json:"updatedAt"`
}

type ClaimStatus uint32

const (
	_ ClaimStatus = iota
	ClaimStatusActive
	ClaimStatusCompleted // user marked as purchased
	ClaimStatusExpired   // expired by store
	ClaimStatusRemoved   // removed by user
)

var _ interface {
	bond.HasBeforeCreate
	bond.HasBeforeUpdate
	bond.HasValidate
} = &Claim{}

var (
	claimStatuses = []string{
		"unknown",
		"active",
		"completed",
		"expired",
	}
)

func (c *Claim) CollectionName() string {
	return `claims`
}

func (c *Claim) BeforeCreate(bond.Session) error {
	c.CreatedAt = GetTimeUTCPointer()
	return nil
}

func (c *Claim) BeforeUpdate(bond.Session) error {
	c.UpdatedAt = GetTimeUTCPointer()
	return nil
}

// TODO: any way to double check promo distance?
func (c *Claim) Validate() error {
	if c.Status == ClaimStatusActive {
		promo, err := DB.Promo.FindByID(c.PromoID)
		if err != nil {
			return err
		}
		//if promo.StartAt != nil && time.Now().Before(*promo.StartAt) {
		//return ErrPromoStart
		//}
		//if promo.EndAt != nil && time.Now().After(*promo.EndAt) {
		//return ErrPromoEnded
		//}
		if promo.PlaceID != c.PlaceID {
			// wrong promo
			return errors.New("mismatch place id")
		}
	}
	return nil
}

// String returns the string value of the status.
func (cs ClaimStatus) String() string {
	return claimStatuses[cs]
}

// MarshalText satisfies TextMarshaler
func (cs ClaimStatus) MarshalText() ([]byte, error) {
	return []byte(cs.String()), nil
}

// UnmarshalText satisfies TextUnmarshaler
func (cs *ClaimStatus) UnmarshalText(text []byte) error {
	enum := string(text)
	for i := 0; i < len(claimStatuses); i++ {
		if enum == claimStatuses[i] {
			*cs = ClaimStatus(i)
			return nil
		}
	}
	return fmt.Errorf("unknown claim status %s", enum)
}
