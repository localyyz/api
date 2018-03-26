package data

import (
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
	_ MerchantApprovalCollection = iota // 0
	MerchantApprovalCollectionSmart
	MerchantApprovalCollectionBoutique
	MerchantApprovalCollectionLuxury
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
	_ MerchantApprovalRejection = iota
	MerchantApprovalRejectionProductQuality
	MerchantApprovalRejectionProductVertical
	MerchantApprovalRejectionReputation
	MerchantApprovalRejectionReturns
	MerchantApprovalRejectionWebsite
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
