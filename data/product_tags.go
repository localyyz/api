package data

import (
	"fmt"
	"time"

	"upper.io/bond"
	db "upper.io/db.v3"
)

type ProductTag struct {
	ID        int64          `db:"id,pk,omitempty" json:"id"`
	ProductID int64          `db:"product_id" json:"productId"`
	PlaceID   int64          `db:"place_id" json:"placeId"`
	Value     string         `db:"value" json:"value"`
	Type      ProductTagType `db:"type" json:"type"`

	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt"`
}

type ProductTagType uint

type ProductTagStore struct {
	bond.Store
}

const (
	_ ProductTagType = iota
	ProductTagTypeGeneral
	ProductTagTypeSize
	ProductTagTypeColor
	ProductTagTypeMaterial
	ProductTagTypeGender
	ProductTagTypeCategory
	ProductTagTypeBrand
)

var (
	productTagTypes = []string{
		"-",
		"general",
		"size",
		"color",
		"material",
		"gender",
		"category",
		"brand",
	}
)

func (t *ProductTag) CollectionName() string {
	return `product_tags`
}

func (store ProductTagStore) FindByProduct(productID int64) ([]*ProductTag, error) {
	return store.FindAll(db.Cond{"product_id": productID})
}

func (store ProductTagStore) FindAll(cond db.Cond) ([]*ProductTag, error) {
	var tags []*ProductTag
	if err := store.Find(cond).All(&tags); err != nil {
		return nil, err
	}
	return tags, nil
}

func (store ProductTagStore) FindOne(cond db.Cond) (*ProductTag, error) {
	var tag *ProductTag
	if err := store.Find(cond).One(&tag); err != nil {
		return nil, err
	}
	return tag, nil
}

// String returns the string value of the status.
func (t ProductTagType) String() string {
	return productTagTypes[t]
}

// MarshalText satisfies TextMarshaler
func (t ProductTagType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalText satisfies TextUnmarshaler
func (t *ProductTagType) UnmarshalText(text []byte) error {
	enum := string(text)
	for i := 0; i < len(productTagTypes); i++ {
		if enum == productTagTypes[i] {
			*t = ProductTagType(i)
			return nil
		}
	}
	return fmt.Errorf("unknown tag type %s", enum)
}
