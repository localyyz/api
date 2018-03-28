package data

import (
	"time"

	"upper.io/bond"
	"upper.io/db.v3"
	"upper.io/db.v3/postgresql"
)

type ProductVariant struct {
	ID        int64 `db:"id,pk,omitempty" json:"id,omitempty"`
	PlaceID   int64 `db:"place_id" json:"placeId"`
	ProductID int64 `db:"product_id" json:"productId"`

	Limits      int64  `db:"limits" json:"limits"`
	Description string `db:"description" json:"description"`
	// external offer id refering to specific variant
	OfferID int64 `db:"offer_id" json:"-"`

	Price     float64           `db:"price" json:"price"`
	PrevPrice float64           `db:"prev_price" json:"prevPrice"`
	Etc       ProductVariantEtc `db:"etc" json:"etc"`

	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt"`
	UpdatedAt *time.Time `db:"updated_at,omitempty" json:"updatedAt"`
	DeletedAt *time.Time `db:"deleted_at,omitempty" json:"deletedAt"`
}

type ProductVariantStore struct {
	bond.Store
}

type ProductVariantEtc struct {
	Price     float64 `json:"prc"`
	PrevPrice float64 `json:"prv"`
	Sku       string  `json:"sku"`

	Size  string `json:"size"`
	Color string `json:"color"`
	*postgresql.JSONBConverter
}

var _ interface {
	bond.HasBeforeCreate
	bond.HasBeforeUpdate
} = &ProductVariant{}

func (p *ProductVariant) CollectionName() string {
	return `product_variants`
}

func (p *ProductVariant) BeforeCreate(sess bond.Session) error {
	// shared checks for updating variant
	if err := p.BeforeUpdate(sess); err != nil {
		return err
	}

	p.UpdatedAt = nil
	p.CreatedAt = GetTimeUTCPointer()

	return nil
}

func (p *ProductVariant) BeforeUpdate(bond.Session) error {
	p.UpdatedAt = GetTimeUTCPointer()

	return nil
}

func (store ProductVariantStore) FindByPlaceID(placeID int64) (*ProductVariant, error) {
	return store.FindOne(db.Cond{"place_id": placeID})
}

func (store ProductVariantStore) FindByID(ID int64) (*ProductVariant, error) {
	return store.FindOne(db.Cond{"id": ID})
}

func (store ProductVariantStore) FindByOfferID(offerID int64) (*ProductVariant, error) {
	return store.FindOne(db.Cond{"offer_id": offerID})
}

func (store ProductVariantStore) FindByProductID(productID int64) ([]*ProductVariant, error) {
	return store.FindAll(db.Cond{"product_id": productID})
}

func (store ProductVariantStore) FindOne(cond db.Cond) (*ProductVariant, error) {
	var variant *ProductVariant
	if err := store.Find(cond).One(&variant); err != nil {
		return nil, err
	}
	return variant, nil
}

func (store ProductVariantStore) FindAll(cond db.Cond) ([]*ProductVariant, error) {
	var variants []*ProductVariant
	if err := store.Find(cond).All(&variants); err != nil {
		return nil, err
	}
	return variants, nil
}
