package data

import (
	"fmt"
	"time"

	"github.com/pressly/lg"
	"upper.io/bond"
	db "upper.io/db.v3"
	"upper.io/db.v3/postgresql"
)

type Place struct {
	ID     int64       `db:"id,pk,omitempty" json:"id,omitempty"`
	Status PlaceStatus `db:"status" json:"status"`

	Name     string      `db:"name" json:"name"`
	Address  string      `db:"address" json:"address"`
	Phone    string      `db:"phone" json:"phone"`
	Website  string      `db:"website" json:"website"`
	ImageURL string      `db:"image_url" json:"imageUrl"`
	Currency string      `db:"currency" json:"currency"`
	Gender   PlaceGender `db:"gender" json:"genderHint"`
	Weight   int32       `db:"weight" json:"weight"`

	ShopifyID string `db:"shopify_id,omitempty" json:"-"`
	Plan      string `db:"shopify_plan,omitempty" json:"-"`
	TOSIP     string `db:"tos_ip" json:"-"`

	// metadata
	IsUsed         bool        `db:"is_used" json:"isUsed"`               // merchant sells used goods
	IsDropShipper  bool        `db:"is_dropshipper" json:"isDropShipper"` // merchant drop shipper
	PlanEnabled    bool        `db:"plan_enabled" json:"planEnabled"`     // if payment plans are enabled for merchant (payment Rollout flag)
	FacebookURL    string      `db:"fb_url" json:"facebookUrl"`
	InstagramURL   string      `db:"instagram_url" json:"instagramUrl"`
	ShippingPolicy PlacePolicy `db:"shipping_policy" json:"shippingPolicy"`
	ReturnPolicy   PlacePolicy `db:"return_policy" json:"returnPolicy"`
	Ratings        PlaceRating `db:"ratings" json:"ratings"`

	PaymentMethods []*PaymentMethod `db:"payment_methods" json:"paymentMethods"`

	CreatedAt   *time.Time `db:"created_at,omitempty" json:"createdAt,omitempty"`
	TOSAgreedAt *time.Time `db:"tos_agreed_at,omitempty" json:"tosAgreedAt,omitempty"`
	ApprovedAt  *time.Time `db:"approved_at,omitempty" json:"approvedAt,omitempty"`
}

// TODO: use go alias
// ie: type PlaceGender = ProductGender
// NOTE: breaks godef https://github.com/rogpeppe/godef/issues/71
// pretty much a deal breaker.. sucks because type aliasing would be amazing here
type PlaceGender ProductGender

type PlaceStatus uint32

type PaymentMethod struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	*postgresql.JSONBConverter
}

type PlaceRating struct {
	Rating        float32 `json:"rating"`
	Count         int64   `json:"count"`
	FBFans        int64   `json:"fbFans"`
	InstFollowers int64   `json:"instFollowers"`
	*postgresql.JSONBConverter
}

type PlacePolicy struct {
	Description string `json:"desc"`
	URL         string `json:"url"`
	*postgresql.JSONBConverter
}

const (
	PlaceStatusUnknown       PlaceStatus = iota // 0
	PlaceStatusWaitAgreement                    // 1
	PlaceStatusWaitApproval                     // 2
	PlaceStatusActive                           // 3
	PlaceStatusInActive                         // 4, closed/dormant
	PlaceStatusReviewing                        // 5, started review process
	PlaceStatusSelectPlan                       // 6, waiting for merchant to select plan
	PlaceStatusRejected                         // 7, rejected
	PlaceStatusUninstalled                      // 8, uninstalled
)

// featured merchant cutoff
const PlaceFeatureWeightCutoff = 5

var _ interface {
	bond.HasBeforeCreate
	bond.HasBeforeUpdate
} = &Place{}

var (
	PlaceGenderUnknown = PlaceGender(ProductGenderUnknown)
	PlaceGenderMale    = PlaceGender(ProductGenderMale)
	PlaceGenderFemale  = PlaceGender(ProductGenderFemale)
	PlaceGenderUnisex  = PlaceGender(ProductGenderUnisex)
)

var (
	placeStatuses = []string{
		"-",
		"waitAgreement",
		"waitApproval",
		"active",
		"inactive",
		"reviewing",
		"selectplan",
		"rejected",
		"uninstalled",
	}
	planWeighting = map[string]int32{
		"":         0,
		"dormant":  0,
		"affliate": 0,
		// above will not have currency modifier

		"starter":      1,
		"basic":        1,
		"custom":       3,
		"professional": 3,
		"unlimited":    5,
		"shopify_plus": 8,
	}
	currencyWeighting = map[string]int32{
		"USD": 1,
		"CAD": 1,
	}
)

func (p *Place) CollectionName() string {
	return `places`
}

func (p *Place) BeforeCreate(sess bond.Session) error {
	p.BeforeUpdate(sess)
	return nil
}

func (p *Place) BeforeUpdate(bond.Session) error {
	if p.Status == PlaceStatusActive {
		w := planWeighting[p.Plan]
		// update place weighting if lower than the plan weighting
		if w > p.Weight {
			p.Weight = w + currencyWeighting[p.Currency]
		}
	}
	return nil
}

/* looks through the priority merchant db if entry of merchantID exists or not*/
func (p *Place) IsPriority() bool {
	exists, err := DB.PriorityMerchant.Find(db.Cond{"place_id": p.ID}).Exists()
	if err != nil {
		lg.Warn("Error: Could not load priority merchant list")
	}
	return exists
}

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

// String returns the string value of the status.
func (s PlaceGender) String() string {
	return productGenders[s]
}

// MarshalText satisfies TextMarshaler
func (s PlaceGender) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalText satisfies TextUnmarshaler
func (s *PlaceGender) UnmarshalText(text []byte) error {
	enum := string(text)
	if gender, ok := productGendersMap[enum]; ok {
		*s = PlaceGender(gender)
		return nil
	}
	return fmt.Errorf("unknown place gender %s", enum)
}
