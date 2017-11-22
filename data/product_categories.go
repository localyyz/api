package data

import (
	"fmt"

	"upper.io/bond"
	db "upper.io/db.v3"
)

type ProductCategory struct {
	ID int64 `db:"id,pk,omitempty" json:"id,omitempty"`

	Type  ProductCategoryType `db:"type" json:"type"`
	Value string              `db:"value" json:"value"`
}

type ProductCategoryStore struct {
	bond.Store
}

type ProductCategoryType uint32

const (
	_                        ProductCategoryType = iota // 0
	ProductCategoryAccessory                            // 1
	ProductCategoryApparel                              // 2
	ProductCategoryHandbag                              // 3
	ProductCategoryJewelry                              // 4
	ProductCategoryShoe                                 // 5
	ProductCategoryCosmetic                             // 6
	ProductCategoryFragrance                            // 7
)

var (
	productCategoryTypes = []string{
		"unknown",
		"accessories",
		"apparel",
		"handbags",
		"jewelry",
		"shoes",
		"cosmetics",
		"fragrances",
	}
)

func (p *ProductCategory) CollectionName() string {
	return `product_categories`
}

func (store ProductCategoryStore) FindByType(t ProductCategoryType) ([]*ProductCategory, error) {
	return store.FindAll(db.Cond{"type": t})
}

func (store ProductCategoryStore) FindOne(cond db.Cond) (*ProductCategory, error) {
	var cat *ProductCategory
	if err := store.Find(cond).One(&cat); err != nil {
		return nil, err
	}
	return cat, nil
}

func (store ProductCategoryStore) FindAll(cond db.Cond) ([]*ProductCategory, error) {
	var cats []*ProductCategory
	if err := store.Find(cond).All(&cats); err != nil {
		return nil, err
	}
	return cats, nil
}

// String returns the string value of the status.
func (t ProductCategoryType) String() string {
	return productCategoryTypes[t]
}

// MarshalText satisfies TextMarshaler
func (t ProductCategoryType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalText satisfies TextUnmarshaler
func (t *ProductCategoryType) UnmarshalText(text []byte) error {
	enum := string(text)
	for i := 0; i < len(productTagTypes); i++ {
		if enum == productCategoryTypes[i] {
			*t = ProductCategoryType(i)
			return nil
		}
	}
	return fmt.Errorf("unknown category type %s", enum)
}
