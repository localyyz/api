package data

import (
	"fmt"
	"time"

	"upper.io/bond"
	"upper.io/db.v2"
)

// TODO: promo should be keyed on placeid and queried on cells
type Promo struct {
	ID         int64     `db:"id,pk,omitempty" json:"id,omitempty"`
	PlaceID    int64     `db:"place_id" json:"placeId"`
	Multiplier int32     `db:"multiplier" json:"multiplier"`
	Type       PromoType `db:"type" json:"type"`

	// Amount of points rewarded
	Reward    int64 `db:"reward" json:"reward"`
	XToReward int64 `db:"x_to_reward" json:"xToReward"` // x amount to complete
	// Duration is the time limit (in seconds) that the
	// promotion must be completed in. -1 for no time limit
	Duration int64 `db:"duration" json:"duration"`

	// TODO: promo status
	StartAt   *time.Time `db:"start_at,omitempty" json:"startAt"`
	EndAt     *time.Time `db:"end_at,omitempty" json:"endAt"`
	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt"`
}

type PromoStore struct {
	bond.Store
}

type PromoType uint32

const (
	_ PromoType = iota
	PromoTypeReachLike
	PromoTypeFirstVisit
	PromoTypeFirstOfDay
)

var (
	promoTypes = []string{
		"-",
		"reach_likes",
		"first_visit",
		"first_ofday",
	}
	PromoDistanceLimit = 200.0
)

func (p *Promo) CollectionName() string {
	return `promos`
}

func (store PromoStore) FindByPlaceID(placeID int64) ([]*Promo, error) {
	return store.FindAll(db.Cond{"place_id": placeID})
}

func (store PromoStore) FindByID(ID int64) (*Promo, error) {
	return store.FindOne(db.Cond{"id": ID})
}

func (store PromoStore) FindOne(cond db.Cond) (*Promo, error) {
	var promo *Promo
	if err := store.Find(cond).One(&promo); err != nil {
		return nil, err
	}
	return promo, nil
}

func (store PromoStore) FindAll(cond db.Cond) ([]*Promo, error) {
	var promos []*Promo
	if err := store.Find(cond).All(&promos); err != nil {
		return nil, err
	}
	return promos, nil
}

// String returns the string value of the status.
func (pt PromoType) String() string {
	return promoTypes[pt]
}

// MarshalText satisfies TextMarshaler
func (pt PromoType) MarshalText() ([]byte, error) {
	return []byte(pt.String()), nil
}

// UnmarshalText satisfies TextUnmarshaler
func (pt *PromoType) UnmarshalText(text []byte) error {
	enum := string(text)
	for i := 0; i < len(promoTypes); i++ {
		if enum == promoTypes[i] {
			*pt = PromoType(i)
			return nil
		}
	}
	return fmt.Errorf("unknown promotion type %s", enum)
}
