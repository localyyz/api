package data

import (
	"fmt"
	"time"

	"upper.io/bond"
	db "upper.io/db.v3"
)

type MerchantApproval struct {
	ID      int64 `db:"id,pk,omitempty" json:"id,omitempty"`
	PlaceID int64 `db:"place_id" json:"placeId"`

	Collection      MerchantApprovalCollection `db:"collection" json:"collection"`
	Category        MerchantApprovalCategory   `db:"category" json:"category"`
	PriceRange      MerchantApprovalPriceRange `db:"price_range" json:"priceRange"`
	RejectionReason MerchantApprovalRejection  `db:"rejection_reason" json:"rejectionReason"`

	CreatedAt *time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt *time.Time `db:"updated_at" json:"updatedAt"`

	ApprovedAt *time.Time `db:"approved_at" json:"approvedAt"`
	RejectedAt *time.Time `db:"rejected_at" json:"rejectedAt"`
}

type MerchantApprovalStore struct {
	bond.Store
}

type MerchantApprovalCollection uint32
type MerchantApprovalCategory uint32
type MerchantApprovalPriceRange uint32
type MerchantApprovalRejection uint32

const (
	_                                  MerchantApprovalCollection = iota // 0
	MerchantApprovalCollectionBoutique                                   // 1
	MerchantApprovalCollectionLuxury                                     // 2
	MerchantApprovalCollectionSmart                                      // 3
)

const (
	_                                              MerchantApprovalCategory = iota // 0
	MerchantApprovalCategoryAccessories                                            // 1
	MerchantApprovalCategoryAthleticOutdoorApparel                                 // 2
	MerchantApprovalCategoryBeautyCosmetics                                        // 3
	MerchantApprovalCategoryConsignment                                            // 4
	MerchantApprovalCategoryElectronics                                            // 5
	MerchantApprovalCategoryFashionApparel                                         // 6
	MerchantApprovalCategoryFashionMultiple                                        // 7
	MerchantApprovalCategoryFootwear                                               // 8
	MerchantApprovalCategoryGeneralStore                                           // 9
	MerchantApprovalCategoryHome                                                   // 10
	MerchantApprovalCategoryIndustrial                                             // 11
	MerchantApprovalCategoryInfant                                                 // 12
	MerchantApprovalCategoryJewelry                                                // 13
	MerchantApprovalCategoryNightwear                                              // 14
	MerchantApprovalCategoryOther                                                  // 15
	MerchantApprovalCategoryPet                                                    // 16
	MerchantApprovalCategoryPreworn                                                // 17
	MerchantApprovalCategorySwimwear                                               // 18
	MerchantApprovalCategoryUnknown                                                // 19
)

const (
	_                                        MerchantApprovalRejection = iota
	MerchantApprovalRejectionProductQuality                            // 1
	MerchantApprovalRejectionProductVertical                           // 2
	MerchantApprovalRejectionReputation                                // 3
	MerchantApprovalRejectionReturns                                   // 4
	MerchantApprovalRejectionWebsite                                   // 5
	MerchantApprovalRejectionInternational                             // 6
)

const (
	_ MerchantApprovalPriceRange = iota
	MerchantApprovalPriceRangeLow
	MerchantApprovalPriceRangeMedium
	MerchantApprovalPriceRangeHigh
)

var _ interface {
	bond.HasBeforeCreate
	bond.HasBeforeUpdate
} = &MerchantApproval{}

var (
	merchantApprovalCategories = []string{
		"-",
		"accessories",
		"athletic/outdoor apparel",
		"beauty/cosmetics",
		"consignment",
		"electronics",
		"fashion apparel",
		"fashion multiple",
		"footwear",
		"general store",
		"home",
		"industrial",
		"infant",
		"jewelry",
		"nightwear",
		"other",
		"pet",
		"preworn",
		"swimwear",
		"unknown",
	}

	merchantApprovalPriceRanges = []string{"-", "low", "medium", "high"}

	merchantApprovalRejectionReasons = []string{
		"unknown",
		"product_quality",
		"reputation",
		"returns",
		"website",
		"international",
	}
)

func (m *MerchantApproval) CollectionName() string {
	return `merchant_approvals`
}

func (m *MerchantApproval) BeforeCreate(sess bond.Session) error {
	if err := m.BeforeUpdate(sess); err != nil {
		return err
	}

	m.UpdatedAt = nil
	m.CreatedAt = GetTimeUTCPointer()

	return nil
}

func (m *MerchantApproval) BeforeUpdate(bond.Session) error {
	m.UpdatedAt = GetTimeUTCPointer()
	return nil
}

func (store MerchantApprovalStore) FindByPlaceID(placeID int64) (*MerchantApproval, error) {
	return store.FindOne(db.Cond{"place_id": placeID})
}

func (store MerchantApprovalStore) FindOne(cond db.Cond) (*MerchantApproval, error) {
	var approval *MerchantApproval
	if err := store.Find(cond).One(&approval); err != nil {
		return nil, err
	}
	return approval, nil
}

// String returns the string value of the status.
func (s MerchantApprovalCategory) String() string {
	return merchantApprovalCategories[s]
}

// MarshalText satisfies TextMarshaler
func (s MerchantApprovalCategory) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalText satisfies TextUnmarshaler
func (s *MerchantApprovalCategory) UnmarshalText(text []byte) error {
	enum := string(text)
	for i := 0; i < len(merchantApprovalCategories); i++ {
		if enum == merchantApprovalCategories[i] {
			*s = MerchantApprovalCategory(i)
			return nil
		}
	}
	return fmt.Errorf("unknown merchant category %s", enum)
}

// String returns the string value of the status.
func (s MerchantApprovalPriceRange) String() string {
	return merchantApprovalPriceRanges[s]
}

// MarshalText satisfies TextMarshaler
func (s MerchantApprovalPriceRange) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalText satisfies TextUnmarshaler
func (s *MerchantApprovalPriceRange) UnmarshalText(text []byte) error {
	enum := string(text)
	for i := 0; i < len(merchantApprovalPriceRanges); i++ {
		if enum == merchantApprovalPriceRanges[i] {
			*s = MerchantApprovalPriceRange(i)
			return nil
		}
	}
	return fmt.Errorf("unknown merchant price range %s", enum)
}

// String returns the string value of the status.
func (s MerchantApprovalRejection) String() string {
	return merchantApprovalRejectionReasons[s]
}

// MarshalText satisfies TextMarshaler
func (s MerchantApprovalRejection) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalText satisfies TextUnmarshaler
func (s *MerchantApprovalRejection) UnmarshalText(text []byte) error {
	enum := string(text)
	for i := 0; i < len(merchantApprovalRejectionReasons); i++ {
		if enum == merchantApprovalRejectionReasons[i] {
			*s = MerchantApprovalRejection(i)
			return nil
		}
	}
	return fmt.Errorf("unknown merchant rejection reason %s", enum)
}
