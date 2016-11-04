package data

import (
	"fmt"
	"time"

	"github.com/goware/geotools"
)

type Place struct {
	ID       int64 `db:"id,pk,omitempty" json:"id,omitempty"`
	LocaleID int64 `db:"locale_id" json:"localeId"`

	Name        string `db:"name" json:"name"`
	Address     string `db:"address" json:"address"`
	Phone       string `db:"phone" json:"phone"`
	Website     string `db:"website" json:"website"`
	Description string `db:"description" json:"description"`
	ImageURL    string `db:"image_url" json:"imageUrl"`

	Gender   PlaceGender `db:"gender" json:"gender"`
	Category Category    `db:"category" json:"category"`

	BlogtoID *int64         `db:"blogto_id,omitempty" json:"-"`
	Geo      geotools.Point `db:"geo" json:"-"`
	Distance float64        `db:"distance,omitempty" json:"distance"` // calculated, not stored in db

	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt,omitempty"`
	UpdatedAt *time.Time `db:"updated_at,omitempty" json:"updatedAt,omitempty"`
}

type PlaceGender uint32

const (
	PlaceGenderUndefined PlaceGender = iota
	PlaceGenderMale
	PlaceGenderFemale
)

var (
	placeGenders = []string{"undefined", "male", "female"}
)

func (p *Place) CollectionName() string {
	return `places`
}

// String returns the string value of the status.
func (s PlaceGender) String() string {
	return placeGenders[s]
}

// MarshalText satisfies TextMarshaler
func (s PlaceGender) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalText satisfies TextUnmarshaler
func (s *PlaceGender) UnmarshalText(text []byte) error {
	enum := string(text)
	for i := 0; i < len(placeGenders); i++ {
		if enum == placeGenders[i] {
			*s = PlaceGender(i)
			return nil
		}
	}
	return fmt.Errorf("unknown place gender %s", enum)
}
