package data

import (
	"fmt"

	"upper.io/bond"
	db "upper.io/db.v3"
)

type BillingPlan struct {
	ID             int64           `db:"id,pk,omitempty" json:"id,omitempty"`
	PlanType       BillingPlanType `db:"plan_type" json:"planType"`
	BillingType    BillingType     `db:"billing_type" json:"billingType"`
	Name           string          `db:"name" json:"name"`
	IsDefault      bool            `db:"is_default" json:"isDefault"`
	Terms          string          `db:"terms" json:"terms"`
	RecurringPrice float64         `db:"recurring_price" json:"recurringPrice"`
	TransationFee  uint32          `db:"transation_fee" json:"transactionFee"`
	CommissionFee  uint32          `db:"commission_fee" json:"commissionFee"`
	OtherFee       uint32          `db:"other_fee" json:"other_fee"`
}

type BillingPlanStore struct {
	bond.Store
}

type BillingPlanType uint32
type BillingType uint32

const (
	_                        BillingPlanType = iota // 0
	BillingPlanTypeStandard                         // 1
	BillingPlanTypeEssential                        // 2
	BillingPlanTypePriority                         // 3
	BillingPlanTypeCustom                           // 4
)

const (
	_                   BillingType = iota // 0
	BillingTypeAnnual                      // 1
	BillingTypeQuaterly                    // 2
	BillingTypeMonthly                     // 3
)

var billingPlanTypes = []string{"-", "standard", "essential", "priority", "custom"}
var billingTypes = []string{"-", "annual", "quaterly", "monthly"}

func (b *BillingPlan) CollectionName() string {
	return `billing_plans`
}

func (store BillingPlanStore) FindByID(ID int64) (*BillingPlan, error) {
	return store.FindOne(db.Cond{"id": ID})
}

func (store BillingPlanStore) FindDefaultByType(planType BillingPlanType) (*BillingPlan, error) {
	return store.FindOne(db.Cond{"plan_type": planType, "is_default": true})
}

func (store BillingPlanStore) FindOne(cond db.Cond) (*BillingPlan, error) {
	var plan *BillingPlan
	if err := store.Find(cond).One(&plan); err != nil {
		return nil, err
	}
	return plan, nil
}

// String returns the string value of the status.
func (b BillingPlanType) String() string {
	return billingPlanTypes[b]
}

// MarshalText satisfies TextMarshaler
func (b BillingPlanType) MarshalText() ([]byte, error) {
	return []byte(b.String()), nil
}

// UnmarshalText satisfies TextUnmarshaler
func (b *BillingPlanType) UnmarshalText(text []byte) error {
	enum := string(text)
	for i := 0; i < len(billingPlanTypes); i++ {
		if enum == billingPlanTypes[i] {
			*b = BillingPlanType(i)
			return nil
		}
	}
	return fmt.Errorf("unknown billing plan type %s", enum)
}

// String returns the string value of the status.
func (b BillingType) String() string {
	return billingTypes[b]
}

// MarshalText satisfies TextMarshaler
func (b BillingType) MarshalText() ([]byte, error) {
	return []byte(b.String()), nil
}

// UnmarshalText satisfies TextUnmarshaler
func (b *BillingType) UnmarshalText(text []byte) error {
	enum := string(text)
	for i := 0; i < len(billingTypes); i++ {
		if enum == billingTypes[i] {
			*b = BillingType(i)
			return nil
		}
	}
	return fmt.Errorf("unknown billing type %s", enum)
}
