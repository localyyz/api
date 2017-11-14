package data

import (
	"upper.io/bond"
	db "upper.io/db.v3"
)

type ProductCategory struct {
	ID int64 `db:"id,pk,omitempty" json:"id,omitempty"`

	Name  string   `db:"name" json:"name"`
	Value []string `db:"value" json:"value"`
}

type ProductCategoryStore struct {
	bond.Store
}

func (p *ProductCategory) CollectionName() string {
	return `product_categories`
}

func (store ProductCategoryStore) FindByName(name string) (*ProductCategory, error) {
	return store.FindOne(db.Cond{"name": name})
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
