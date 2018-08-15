package data

import (
	"fmt"

	"upper.io/bond"
	db "upper.io/db.v3"
)

type Category struct {
	Type     CategoryType  `db:"type" json:"type"`
	Value    string        `db:"value" json:"value"`
	Mapping  string        `db:"mapping" json:"mapping"`
	ImageURL string        `db:"image_url" json:"imageUrl"`
	Gender   ProductGender `db:"gender" json:"-"`
	Weight   int32         `db:"weight" json:"-"`
}

type CategoryStore struct {
	bond.Store
}

type CategoryType uint32

const (
	_                 CategoryType = iota // 0
	CategoryAccessory                     // 1
	CategoryApparel                       // 2
	CategoryHandbag                       // 3
	CategoryJewelry                       // 4
	CategoryShoe                          // 5
	CategoryCosmetic                      // 6
	CategoryFragrance                     // 7
	CategoryHome                          // 8
	CategoryBag                           // 9
	CategoryLingerie                      // 10
	CategorySneaker                       // 11
	CategorySwimwear                      // 12

	// Special non DB category
	CategorySale  // 13
	CategoryNewIn // 14
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
	}

	Categories = []CategoryType{
		CategoryAccessory,
		CategoryApparel,
		CategoryHandbag,
		CategoryJewelry,
		CategoryShoe,
		CategoryCosmetic,
		CategoryFragrance,
		CategoryHome,
		CategoryBag,
		CategoryLingerie,
		CategorySneaker,
		CategorySwimwear,
		CategorySale,
		CategoryNewIn,
	}
)

func (p *Category) CollectionName() string {
	return `product_categories`
}

func (store CategoryStore) FindByType(t CategoryType) ([]*Category, error) {
	return store.FindAll(db.Cond{"type": t})
}

func (store CategoryStore) FindByMapping(v string) ([]*Category, error) {
	return store.FindAll(db.Cond{"mapping": v})
}

func (store CategoryStore) FindOne(cond db.Cond) (*Category, error) {
	var cat *Category
	if err := store.Find(cond).One(&cat); err != nil {
		return nil, err
	}
	return cat, nil
}

func (store CategoryStore) FindAll(cond db.Cond) ([]*Category, error) {
	var cats []*Category
	if err := store.Find(cond).All(&cats); err != nil {
		return nil, err
	}
	return cats, nil
}

// String returns the string value of the status.
func (t CategoryType) String() string {
	return categoryTypes[t]
}

// MarshalText satisfies TextMarshaler
func (t CategoryType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalText satisfies TextUnmarshaler
func (t *CategoryType) UnmarshalText(text []byte) error {
	enum := string(text)
	for i := 0; i < len(categoryTypes); i++ {
		if enum == categoryTypes[i] {
			*t = CategoryType(i)
			return nil
		}
	}
	return fmt.Errorf("unknown category type %s", enum)
}
