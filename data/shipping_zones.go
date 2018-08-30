package data

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"upper.io/bond"
	db "upper.io/db.v3"
)

type ShippingZone struct {
	ID         int64 `db:"id,pk,omitempty" json:"id,omitempty"`
	PlaceID    int64 `db:"place_id" json:"placeID"`
	ExternalID int64 `db:"external_id" json:"externalID"`

	Type        ShippingZoneType `db:"type" json:"type"`
	Name        string           `json:"name" db:"name"`
	Description string           `json:"description" db:"description"`
	Price       float64          `json:"price" db:"price"`
	Country     string           `json:"country" db:"country"`
	Regions     Regions          `json:"regions" db:"regions"`
	// Regions is inserted as "<name>:<code>"

	// NOTE: for now, just US + canada
	// TODO: expand to applicable country

	WeightLow    float64 `json:"weightLow" db:"weight_low"`
	WeightHigh   float64 `json:"weightHigh" db:"weight_high"`
	SubtotalLow  float64 `json:"subtotalLow" db:"subtotal_low"`
	SubtotalHigh float64 `json:"subtotalHigh" db:"subtotal_high"`
}
type ShippingZoneType uint32

const (
	_ ShippingZoneType = iota
	ShippingZoneTypeByWeight
	ShippingZoneTypeByPrice
	ShippingZoneTypeByCarrier
)

var (
	ErrRegionDecode = errors.New("region decode")

	shippingZoneTypes = []string{
		"-",
		"weight",
		"price",
		"carrier",
	}
)

const RegionSeparator = ":"

type Region struct {
	Region     string
	RegionCode string
}

type Regions []Region

func (r Region) String() string {
	return strings.ToLower(fmt.Sprintf("%s:%s", r.Region, r.RegionCode))
}

func (r *Region) UnmarshalJSON(b []byte) error {
	regions := bytes.Split(b, []byte{58})
	if len(regions) < 2 {
		return ErrRegionDecode
	}
	r.Region = string(regions[0])
	r.RegionCode = string(regions[1])
	return nil
}

func (r Regions) MarshalJSON() ([]byte, error) {
	s := make([]string, len(r))
	for i, v := range r {
		s[i] = v.String()
	}
	return json.Marshal(s)
}

type ShippingZoneStore struct {
	bond.Store
}

func (p *ShippingZone) CollectionName() string {
	return `shipping_zones`
}

func (store ShippingZoneStore) FindByPlaceID(placeID int64) ([]*ShippingZone, error) {
	return store.FindAll(db.Cond{"place_id": placeID})
}

func (store ShippingZoneStore) FindAll(cond db.Cond) ([]*ShippingZone, error) {
	var zones []*ShippingZone
	if err := store.Find(cond).All(&zones); err != nil {
		return nil, err
	}
	return zones, nil
}

func (store ShippingZoneStore) FindOne(cond db.Cond) (*ShippingZone, error) {
	var zone *ShippingZone
	if err := store.Find(cond).One(&zone); err != nil {
		return nil, err
	}
	return zone, nil
}

// String returns the string value of the status.
func (s ShippingZoneType) String() string {
	return shippingZoneTypes[s]
}

// MarshalText satisfies TextMarshaler
func (s ShippingZoneType) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalText satisfies TextUnmarshaler
func (s *ShippingZoneType) UnmarshalText(text []byte) error {
	enum := string(text)
	for i := 0; i < len(placeStatuses); i++ {
		if enum == shippingZoneTypes[i] {
			*s = ShippingZoneType(i)
			return nil
		}
	}
	return fmt.Errorf("unknown shipping zone type status %s", enum)
}
