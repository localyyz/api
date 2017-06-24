package data

import (
	"fmt"
	"time"

	"upper.io/bond"
	db "upper.io/db.v3"
)

type Product struct {
	ID         int64  `db:"id,pk,omitempty" json:"id,omitempty"`
	PlaceID    int64  `db:"place_id" json:"placeId"`
	ExternalID string `db:"external_id" json:"-"`

	Title       string     `db:"title" json:"title"`
	Description string     `db:"description" json:"description"`
	ImageUrl    string     `db:"image_url" json:"imageUrl"`
	Etc         ProductEtc `db:"etc,jsonb" json:"etc"`

	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt"`
	UpdatedAt *time.Time `db:"updated_at,omitempty" json:"updatedAt"`
	DeletedAt *time.Time `db:"deleted_at,omitempty" json:"deletedAt"`
}

type ProductEtc struct {
	Brand string `json:"brand"`
	Type  string `json:"type"`
}

type ProductStore struct {
	bond.Store
}

func (p *Product) CollectionName() string {
	return `products`
}

// TODO parse item title into tags
func (store ProductStore) Fuzzy(q string) ([]*Product, error) {
	tags, err := DB.ProductTag.FindAll(db.Cond{
		"value ~*": fmt.Sprint("\\m(", q, ")"),
	})
	if err != nil {
		return nil, err
	}

	byTitle, err := store.FindAll(db.Cond{
		"title ~*": fmt.Sprint("\\m(", q, ")"),
	})
	if err != nil {
		return nil, err
	}

	productIDs := make([]int64, len(tags))
	for i, t := range tags {
		productIDs[i] = t.ProductID
	}

	products, err := store.FindAll(db.Cond{"id": productIDs})
	if err != nil {
		return nil, err
	}

	products = append(products, byTitle...)
	return products, nil
}

func (store ProductStore) FindPromos(productID int64) ([]*Promo, error) {
	return DB.Promo.FindAll(db.Cond{
		"product_id": productID,
		"status":     PromoStatusActive,
	})
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
