package data

import (
	"fmt"
	"time"

	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"github.com/goware/geotools"
	"upper.io/db.v3/postgresql"
)

type Place struct {
	ID       int64       `db:"id,pk,omitempty" json:"id,omitempty"`
	LocaleID int64       `db:"locale_id" json:"localeId"`
	Status   PlaceStatus `db:"status" json:"status"`

	Name        string      `db:"name" json:"name"`
	Address     string      `db:"address" json:"address"`
	Phone       string      `db:"phone" json:"phone"`
	Website     string      `db:"website" json:"website"`
	Description string      `db:"description" json:"description"`
	ImageURL    string      `db:"image_url" json:"imageUrl"`
	Currency    string      `db:"currency" json:"currency"`
	Gender      PlaceGender `db:"gender" json:"genderHint"`
	Weight      int32       `db:"weight" json:"weight"`

	ShopifyID string         `db:"shopify_id,omitempty" json:"-"`
	Geo       geotools.Point `db:"geo" json:"-"`
	Distance  float64        `db:"distance,omitempty" json:"distance"` // calculated, not stored in db
	TOSIP     string         `db:"tos_ip" json:"-"`

	Billing PlaceBilling `db:"billing" json:"billing"`

	CreatedAt   *time.Time `db:"created_at,omitempty" json:"createdAt,omitempty"`
	UpdatedAt   *time.Time `db:"updated_at,omitempty" json:"updatedAt,omitempty"`
	TOSAgreedAt *time.Time `db:"tos_agreed_at,omitempty" json:"tosAgreedAt,omitempty"`
	ApprovedAt  *time.Time `db:"approved_at,omitempty" json:"approvedAt,omitempty"`
}

// TODO: use go alias
// ie: type PlaceGender = ProductGender
// NOTE: breaks godef https://github.com/rogpeppe/godef/issues/71
// pretty much a deal breaker.. sucks because type aliasing would be amazing here
type PlaceGender ProductGender

var (
	PlaceGenderUnknown = PlaceGender(ProductGenderUnknown)
	PlaceGenderMale    = PlaceGender(ProductGenderMale)
	PlaceGenderFemale  = PlaceGender(ProductGenderFemale)
	PlaceGenderUnisex  = PlaceGender(ProductGenderUnisex)
)

func (p *Place) CollectionName() string {
	return `places`
}

type PlaceBilling struct {
	*shopify.Billing
	*postgresql.JSONBConverter
}

type PlaceStatus uint32

const (
	PlaceStatusUnknown       PlaceStatus = iota // 0
	PlaceStatusWaitAgreement                    // 1
	PlaceStatusWaitApproval                     // 2
	PlaceStatusActive                           // 3
	PlaceStatusInActive                         // 4
)

var (
	placeStatuses = []string{
		"-",
		"waitAgreement",
		"waitApproval",
		"active",
		"inactive",
	}
)

// String returns the string value of the status.
func (s PlaceStatus) String() string {
	return placeStatuses[s]
}

// MarshalText satisfies TextMarshaler
func (s PlaceStatus) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalText satisfies TextUnmarshaler
func (s *PlaceStatus) UnmarshalText(text []byte) error {
	enum := string(text)
	for i := 0; i < len(placeStatuses); i++ {
		if enum == placeStatuses[i] {
			*s = PlaceStatus(i)
			return nil
		}
	}
	return fmt.Errorf("unknown place status %s", enum)
}
