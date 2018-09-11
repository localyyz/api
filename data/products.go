package data

import (
	"fmt"
	"time"

	"upper.io/bond"
	db "upper.io/db.v3"
	"upper.io/db.v3/postgresql"
)

type Product struct {
	ID      int64 `db:"id,pk,omitempty" json:"id,omitempty"`
	PlaceID int64 `db:"place_id" json:"placeId"`

	Title       string        `db:"title" json:"title"`
	Description string        `db:"description" json:"description"`
	Brand       string        `db:"brand" json:"brand"`
	Gender      ProductGender `db:"gender" json:"genderHint"`
	Score       int64         `db:"score" json:"score"`

	CategoryID *int64          `db:"category_id" json:"categoryId"`
	Category   ProductCategory `db:"category" json:"category"`
	Status     ProductStatus   `db:"status" json:"status"`

	// price from variants
	Price        float64    `db:"price" json:"price"`
	DiscountPct  float64    `db:"discount_pct" json:"discountPct"`
	DiscountedAt *time.Time `db:"discounted_at,omitempty" json:"discountedAt"`

	// external id
	ExternalID     *int64 `db:"external_id,omitempty" json:"-"`
	ExternalHandle string `db:"external_handle" json:"-"`

	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt"`
	UpdatedAt *time.Time `db:"updated_at,omitempty" json:"updatedAt"`
	DeletedAt *time.Time `db:"deleted_at,omitempty" json:"deletedAt"`
}

type ProductCategory struct {
	Type  ProductCategoryType `json:"type,omitempty"`
	Value string              `json:"value,omitempty"`
	*postgresql.JSONBConverter
}

var productStatuses = []string{
	"-",
	"pending",
	"processing",
	"approved",
	"rejected",
	"deleted",
	"outstock",
}

type ProductStatus uint32

const (
	_                       ProductStatus = iota //0
	ProductStatusPending                         //1
	ProductStatusProcessing                      //2
	ProductStatusApproved                        //3
	ProductStatusRejected                        //4
	ProductStatusDeleted                         //5
	ProductStatusOutofStock                      //6
)

type ProductGender uint32

const (
	ProductGenderUnknown ProductGender = iota
	ProductGenderMale
	ProductGenderFemale
	ProductGenderUnisex
	ProductGenderKid
)

var productGenders = []string{
	"-",
	"man",
	"woman",
	"unisex",
	"kid",
}

const (
	// NOTE on magic numbers
	//
	// ranking type 32 => normalizes the rank with `x / (1+x)` where x is the original rank
	// modifier 4 => is the top 70th (another magical number) of our merchant
	//      weights greater than 0
	ProductQueryWeight       = `ts_rank_cd(tsv, to_tsquery($$?$$), 32) + (p.score / (4+p.score::float))`
	ProductQueryWeightWithID = `ts_rank_cd(tsv, to_tsquery($$?$$), 32) + (ln(p.id)+p.score) / (4+ln(p.id)+p.score::float)`
	ProductFuzzyWeight       = `CASE WHEN category != '{}' THEN 1 ELSE 0 END + ts_rank_cd(tsv, to_tsquery(?), 16) + (p.score / (4+p.score::float))`
	ProductFuzzyWeightWithID = `CASE WHEN category != '{}' THEN 1 ELSE 0 END + ts_rank_cd(tsv, to_tsquery(?), 16) + ((ln(p.id)+p.score) / (4+ln(p.id)+p.score::float))`
)

type ProductCategoryType uint32

const (
	_                 ProductCategoryType = iota // 0
	CategoryAccessory                            // 1
	CategoryApparel                              // 2
	CategoryHandbag                              // 3
	CategoryJewelry                              // 4
	CategoryShoe                                 // 5
	CategoryCosmetic                             // 6
	CategoryFragrance                            // 7
	CategoryHome                                 // 8
	CategoryBag                                  // 9
	CategoryLingerie                             // 10
	CategorySneaker                              // 11
	CategorySwimwear                             // 12

	// Special non DB category
	CategorySale       // 13
	CategoryNewIn      // 14
	CategoryCollection // 15
)

var (
	categoryTypes = []string{
		"unknown",
		"accessories",
		"apparel",
		"handbags",
		"jewelry",
		"shoes",
		"cosmetics",
		"fragrances",
		"home",
		"bags",
		"lingerie",
		"sneakers",
		"swimwear",
		"sales",
		"newin",
		"collections",
	}
)

type ProductStore struct {
	bond.Store
}

var _ interface {
	bond.HasBeforeCreate
	bond.HasBeforeUpdate
} = &Product{}

func (p *Product) CollectionName() string {
	return `products`
}

func (p *Product) BeforeCreate(sess bond.Session) error {
	// shared checks for updating variant
	if err := p.BeforeUpdate(sess); err != nil {
		return err
	}
	p.UpdatedAt = nil

	if p.CreatedAt == nil {
		p.CreatedAt = GetTimeUTCPointer()
	}
	return nil
}

func (p *Product) BeforeUpdate(bond.Session) error {
	p.UpdatedAt = GetTimeUTCPointer()

	return nil
}

func (store ProductStore) FindByID(ID int64) (*Product, error) {
	return store.FindOne(db.Cond{"id": ID})
}

func (store ProductStore) FindByExternalID(extID int64) (*Product, error) {
	return store.FindOne(db.Cond{"external_id": extID})
}

func (store ProductStore) FindOne(cond db.Cond) (*Product, error) {
	var product *Product
	if err := store.Find(cond).One(&product); err != nil {
		return nil, err
	}
	return product, nil
}

func (store ProductStore) FindAll(cond db.Cond) ([]*Product, error) {
	var products []*Product
	if err := store.Find(cond).All(&products); err != nil {
		return nil, err
	}
	return products, nil
}

type FeatureProduct struct {
	ProductID     int64      `db:"product_id"`
	Ordering      uint32     `db:"ordering"`
	ImageUrl      string     `db:"image_url"`
	FeaturedAt    *time.Time `db:"featured_at"`
	EndFeaturedAt *time.Time `db:"end_featured_at"`
}

type FeatureProductStore struct {
	bond.Store
}

func (FeatureProduct) CollectionName() string {
	return `feature_products`
}

// String returns the string value of the status.
func (s ProductGender) String() string {
	return productGenders[s]
}

// MarshalText satisfies TextMarshaler
func (s ProductGender) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalText satisfies TextUnmarshaler
func (s *ProductGender) UnmarshalText(text []byte) error {
	enum := string(text)
	for i := 0; i < len(productGenders); i++ {
		if enum == productGenders[i] {
			*s = ProductGender(i)
			return nil
		}
	}
	return fmt.Errorf("unknown product gender %s", enum)
}

// String returns the string value of the status.
func (s ProductStatus) String() string {
	return productStatuses[s]
}

// MarshalText satisfies TextMarshaler
func (s ProductStatus) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalText satisfies TextUnmarshaler
func (s *ProductStatus) UnmarshalText(text []byte) error {
	enum := string(text)
	for i := 0; i < len(productStatuses); i++ {
		if enum == productStatuses[i] {
			*s = ProductStatus(i)
			return nil
		}
	}
	return fmt.Errorf("unknown product status %s", enum)
}

// String returns the string value of the status.
func (t ProductCategoryType) String() string {
	return categoryTypes[t]
}

// MarshalText satisfies TextMarshaler
func (t ProductCategoryType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalText satisfies TextUnmarshaler
func (t *ProductCategoryType) UnmarshalText(text []byte) error {
	enum := string(text)
	for i := 0; i < len(categoryTypes); i++ {
		if enum == categoryTypes[i] {
			*t = ProductCategoryType(i)
			return nil
		}
	}
	return fmt.Errorf("unknown product category type %s", enum)
}
