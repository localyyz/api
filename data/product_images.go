package data

import (
	"upper.io/bond"
	db "upper.io/db.v3"
)

type ProductImage struct {
	ID         int64  `db:"id,pk,omitempty" json:"id,omitempty"`
	ProductID  int64  `db:"product_id" json:"productId,omitempty"`
	ExternalID int64  `db:"external_id,omitempty" json:"-"`
	ImageURL   string `db:"image_url" json:"imageUrl"`
	Ordering   int32  `db:"ordering" json:"ordering"`

	VariantIDs []int64 `db:"-" json:"-"`
}

type ProductImageStore struct {
	bond.Store
}

func (p *ProductImage) CollectionName() string {
	return `product_images`
}

func (store ProductImageStore) FindByID(ID int64) (*ProductImage, error) {
	return store.FindOne(db.Cond{"id": ID})
}

func (store ProductImageStore) FindByExternalID(extID int64) (*ProductImage, error) {
	return store.FindOne(db.Cond{"external_id": extID})
}

func (store ProductImageStore) FindByProductID(productID int64) ([]*ProductImage, error) {
	return store.FindAll(db.Cond{"product_id": productID})
}

func (store ProductImageStore) FindOne(cond db.Cond) (*ProductImage, error) {
	var image *ProductImage
	if err := store.Find(cond).One(&image); err != nil {
		return nil, err
	}
	return image, nil
}

func (store ProductImageStore) FindAll(cond db.Cond) ([]*ProductImage, error) {
	var images []*ProductImage
	if err := store.Find(cond).OrderBy("ordering").All(&images); err != nil {
		return nil, err
	}
	return images, nil
}

/* variant image pivot table */

type VariantImage struct {
	VariantID int64 `db:"variant_id" json:"-"`
	ImageID   int64 `db:"image_id" json:"-"`
}

type VariantImageStore struct {
	bond.Store
}

func (v *VariantImage) CollectionName() string {
	return `variant_images_pivot`
}

func (store VariantImageStore) FindByVariantIDs(variantIDs ...int64) ([]*VariantImage, error) {
	return store.FindAll(db.Cond{"variant_id": variantIDs})
}

func (store VariantImageStore) FindAll(cond db.Cond) ([]*VariantImage, error) {
	var images []*VariantImage
	if err := store.Find(cond).All(&images); err != nil {
		return nil, err
	}
	return images, nil
}

func (store VariantImageStore) FindByVariantID(variantID int64) (*VariantImage, error) {
	var image *VariantImage
	if err := store.Find(db.Cond{"variant_id": variantID}).One(&image); err != nil {
		return nil, err
	}
	return image, nil
}
