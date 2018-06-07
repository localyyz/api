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

	Category ProductCategory `db:"category" json:"category"`
	Status   ProductStatus   `db:"status" json:"status"`

	// price from variants
	Price       float64 `db:"price" json:"price"`
	DiscountPct float64 `db:"discount_pct" json:"discount_pct"`

	// external id
	ExternalID     *int64 `db:"external_id,omitempty" json:"-"`
	ExternalHandle string `db:"external_handle" json:"-"`

	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt"`
	UpdatedAt *time.Time `db:"updated_at,omitempty" json:"updatedAt"`
	DeletedAt *time.Time `db:"deleted_at,omitempty" json:"deletedAt"`
}

type ProductCategory struct {
	Type  CategoryType `json:"type,omitempty"`
	Value string       `json:"value,omitempty"`
	*postgresql.JSONBConverter
}

var productStatuses = []string{"-", "pending", "processing", "approved", "rejected"}

type ProductStatus uint32

const (
	_                       ProductStatus = iota //0
	ProductStatusPending                         //1
	ProductStatusProcessing                      //2
	ProductStatusApproved                        //3
	ProductStatusRejected                        //4
)

type ProductGender uint32

const (
	ProductGenderUnknown ProductGender = iota
	ProductGenderMale
	ProductGenderFemale
	ProductGenderUnisex
)

const (
	ProductQueryWeight       = `ts_rank_cd(tsv, to_tsquery($$?$$), 32) + (p.score / (4+p.score::float)) as _rank`
	ProductQueryWeightWithID = `ts_rank_cd(tsv, to_tsquery($$?$$), 32) + (ln(p.id)+p.score) / (4+ln(p.id)+p.score::float) as _rank`
	ProductFuzzyWeight       = `CASE WHEN category != '{}' THEN 1 ELSE 0 END + ts_rank_cd(tsv, to_tsquery(?), 16) + (p.score / (4+p.score::float)) as _rank`
	ProductFuzzyWeightWithID = `CASE WHEN category != '{}' THEN 1 ELSE 0 END + ts_rank_cd(tsv, to_tsquery(?), 16) + ((ln(p.id)+p.score) / (4+ln(p.id)+p.score::float)) as _rank`
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

var productGenders = []string{"-", "man", "woman", "unisex"}

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
