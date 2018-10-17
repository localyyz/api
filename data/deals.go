package data

import (
	"time"
	"upper.io/db.v3/postgresql"

	"github.com/pkg/errors"
	"upper.io/bond"
	db "upper.io/db.v3"
)

type Deal struct {
	ID         int64      `db:"id,pk,omitempty" json:"id"`
	Status     DealStatus `db:"status" json:"status"`
	ExternalID int64      `db:"external_id" json:"externalId"`
	MerchantID int64      `db:"merchant_id" json:"merchantId"`
	ImageURL   string  `db:"image_url" json:"imageUrl"`

	// parent deal id in which this deal inherits from
	// used in conjunction with userID to specify an
	// user specific deal.
	ParentID        *int64          `db:"parent_id,omitempty" json:"parentId,omitempty"`
	UserID          *int64          `db:"user_id,omitempty" json:"userId,omitempty"`
	Type            DealType        `db:"type" json:"dealType"`
	ProductListType ProductListType `db:"product_type" json:"productType"`

	Code             string               `db:"code" json:"code"`
	Value            float64              `db:"value" json:"value"`
	UsageLimit       int32                `db:"use_limit" json:"useLimit"`
	OncePerCustomer  bool                 `db:"use_once" json:"useOnce"`
	Timed            bool                 `db:"timed" json:"timed"`
	Featured		 bool				  `db:"featured" json:"featured"`
	Prerequisite     DealPrerequisite     `db:"prerequisite" json"prerequisite"`
	BXGYPrerequisite BXGYDealPrerequisite `db:"bxgy" json"bxgy"`

	StartAt *time.Time `db:"start_at,omitempty" json:"startAt"`
	EndAt   *time.Time `db:"end_at,omitempty" json:"endAt"`
}

type DealStore struct {
	bond.Store
}

type DealStatus uint32
type DealType uint32
type ProductListType uint32

const (
	_                  DealStatus = iota // 0
	DealStatusQueued                     // 1
	DealStatusInactive                   // 2
	DealStatusActive                     // 3
)

const (
	_                     DealType = iota // 0
	DealTypeAmountOff                     // 1
	DealTypePercentageOff                 // 2
	DealTypeFreeShipping                  // 3
	DealTypeBXGY                          // 4
)

const (
	ProductListTypeUnKnown    ProductListType = iota //0
	ProductListTypeAssociated                        //1
	ProductListTypeBXGY                              //2
	ProductListTypeMerch                             //3
)

type DealPrerequisite struct {
	ShippingPriceRange int `json:"shipping_price_range,omitempty"`
	SubtotalRange      int `json:"subtotal_range,omitempty"`
	QuantityRange      int `json:"quantity_range,omitempty"`

	*postgresql.JSONBConverter
}

type BXGYDealPrerequisite struct {
	EntitledProductIds []int64 `json:"entitled_product_ids,omitempty"`

	PrerequisiteProductIds []int64 `json:"prerequisite_product_ids,omitempty"`

	//Quantities for the BXGY ratio
	PrerequisiteQuantityBXGY int `json:"prerequisite_quantity,omitempty"`
	EntitledQuantityBXGY     int `json:"entitled_quantity,omitempty"`

	AllocationLimit int `json:"allocation_limit,omitempty"`

	*postgresql.JSONBConverter
}

func (d *Deal) CollectionName() string {
	return `deals`
}

func (store DealStore) FindByID(ID int64) (*Deal, error) {
	return store.FindOne(db.Cond{"id": ID})
}
func (store DealStore) FindByPlace(ID int64) ([]*Deal, error) {
	var ds []*Deal
	if err := store.Find(db.Cond{"merchant_id": ID}).All(&ds); err != nil {
		return nil, err
	}
	return ds, nil
}

func (store DealStore) FindAll(cond db.Cond) ([]*Deal, error) {
	var ds []*Deal
	if err := store.Find(cond).All(&ds); err != nil {
		return nil, err
	}
	return ds, nil
}

func (store DealStore) FindOne(cond db.Cond) (*Deal, error) {
	var d *Deal
	if err := store.Find(cond).One(&d); err != nil {
		return nil, err
	}
	return d, nil
}

/*
	Returns the total number of successfull checkouts of a collection
*/
func (d *Deal) GetCheckoutCount() (int, error) {
	row, err := DB.Select(db.Raw("count(1) as _t")).
		From("deal_products as dp").
		LeftJoin("cart_items as ci").On("dp.product_id = ci.product_id").
		LeftJoin("carts c").On("c.id = ci.cart_id").
		Where(
			db.Cond{
				"dp.deal_id": d.ID,
				"c.status":   CartStatusPaymentSuccess,
			},
		).QueryRow()
	if err != nil {
		return 0, errors.Wrap(err, "collection checkout prepare")
	}

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, errors.Wrap(err, "collection checkout scan")
	}

	return count, nil
}

// deal product

type DealProduct struct {
	DealID    int64 `db:"deal_id"`
	ProductID int64 `db:"product_id"`
}

type DealProductStore struct {
	bond.Store
}

func (d *DealProduct) CollectionName() string {
	return `deal_products`
}

func (store DealProductStore) FindByDealID(dealID int64) ([]*DealProduct, error) {
	return store.FindAll(db.Cond{"deal_id": dealID})
}

func (store DealProductStore) FindAll(cond db.Cond) ([]*DealProduct, error) {
	var dps []*DealProduct
	if err := store.Find(cond).All(&dps); err != nil {
		return nil, err
	}
	return dps, nil
}

func (store DealProductStore) FindOne(cond db.Cond) (*DealProduct, error) {
	var dp *DealProduct
	if err := store.Find(cond).One(&dp); err != nil {
		return nil, err
	}
	return dp, nil
}
