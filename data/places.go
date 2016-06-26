package data

import (
	"fmt"
	"time"
)

// For now, just map places to google id
// NOTE: we would have to carry this data on our side eventually
type Place struct {
	ID int64 `db:"id,pk,omitempty" json:"id,omitempty"`
	// external google places id
	GoogleID       string `db:"google_id" json:"googleId"`
	NeighborhoodID int64  `db:"neighborhood_id" json:"neighborhoodId"`

	Type    PlaceType `db:"place_type" json:"place_type"`
	Address string    `db:"address" json:"address"`
	Phone   string    `db:"phone" json:"phone"`
	Website string    `db:"website" json:"website"`

	Etc *PlaceEtc `db:"etc,jsonb" json:"etc"`

	// TODO: Hours
	// TODO: Subtypes?

	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt,omitempty"`
	UpdatedAt *time.Time `db:"updated_at,omitempty" json:"updatedAt,omitempty"`
}

type Neighborhood struct {
	ID   int64  `db:"id,pk,omitempty" json:"id,omitempty"`
	Name string `db:"name" json:"name"`
}

type PlaceType uint

const (
	PlaceTypeUnknown PlaceType = iota
	PlaceTypeFood
	PlaceTypeNightLife
	PlaceTypeShopping
	PlaceTypeEntertainment
)

var placeTypes = []string{"unknown", "food", "night life", "shopping", "entertainment"}

type PlaceEtc struct {
	Latitude     float64  `json:"latitude"`
	Longitude    float64  `json:"longitude"`
	GoogleRating float64  `json:"google_rating"`
	GoogleTypes  []string `json:"google_types"`
}

func (p *Place) CollectionName() string {
	return `places`
}

func (n *Neighborhood) CollectionName() string {
	return `neighborhoods`
}

// String returns the string value of the status.
func (pt PlaceType) String() string {
	return placeTypes[pt]
}

// MarshalText satisfies TextMarshaler
func (pt PlaceType) MarshalText() ([]byte, error) {
	return []byte(pt.String()), nil
}

// UnmarshalText satisfies TextUnmarshaler
func (pt *PlaceType) UnmarshalText(text []byte) error {
	enum := string(text)
	for i := 0; i < len(placeTypes); i++ {
		if enum == placeTypes[i] {
			*pt = PlaceType(i)
			return nil
		}
	}
	return fmt.Errorf("unknown place type %s", enum)
}
