package data

import (
	"fmt"

	"upper.io/bond"
	db "upper.io/db.v3"
	"upper.io/db.v3/postgresql"
)

// category white list
type Whitelist struct {
	CategoryID *int64        `db:"category_id,omitempty"`
	Value      string        `db:"value"`
	Gender     ProductGender `db:"gender"`
	Weight     int32         `db:"weight"`

	// special whitelist tags
	IsSpecial bool `db:"-"`
	IsIgnore  bool `db:"-"`

	// legacy columns
	Type ProductCategoryType `db:"type"`
}

type WhitelistStore struct {
	bond.Store
}

func (*Whitelist) CollectionName() string {
	return `whitelist`
}

func (store WhitelistStore) FindAll(cond db.Cond) ([]*Whitelist, error) {
	var list []*Whitelist
	if err := store.Find(cond).All(&list); err != nil {
		return nil, err
	}
	return list, nil
}

type ProductCategoryType uint32

type ProductCategory struct {
	Type  ProductCategoryType `json:"type,omitempty"`
	Value string              `json:"value,omitempty"`
	*postgresql.JSONBConverter
}

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
