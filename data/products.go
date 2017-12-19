package data

import (
	"time"

	"upper.io/bond"
	db "upper.io/db.v3"
	"upper.io/db.v3/postgresql"
)

type Product struct {
	ID      int64 `db:"id,pk,omitempty" json:"id,omitempty"`
	PlaceID int64 `db:"place_id" json:"placeId"`

	Title       string     `db:"title" json:"title"`
	Description string     `db:"description" json:"description"`
	ImageUrl    string     `db:"image_url" json:"imageUrl"`
	Etc         ProductEtc `db:"etc" json:"etc"`

	// external id
	ExternalID     *int64 `db:"external_id,omitempty" json:"-"`
	ExternalHandle string `db:"external_handle" json:"-"`

	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt"`
	UpdatedAt *time.Time `db:"updated_at,omitempty" json:"updatedAt"`
	DeletedAt *time.Time `db:"deleted_at,omitempty" json:"deletedAt"`
}

type ProductEtc struct {
	Brand  string   `json:"brand"`
	Type   string   `json:"type"`
	Images []string `json:"images"`
	*postgresql.JSONBConverter
}

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
	p.CreatedAt = GetTimeUTCPointer()
	return nil
}

func (p *Product) BeforeUpdate(bond.Session) error {
	p.UpdatedAt = GetTimeUTCPointer()

	return nil
}

func (store ProductStore) FindByID(ID int64) (*Product, error) {
	return store.FindOne(db.Cond{"id": ID})
}

func (store ProductStore) FindByExternalID(extID string) (*Product, error) {
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
