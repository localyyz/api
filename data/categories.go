package data

import (
	"upper.io/bond"
	db "upper.io/db.v3"
)

type Category struct {
	ID       int64  `json:"id" db:"id"`
	Value    string `json:"value" db:"value"`
	Label    string `json:"label" db:"label"`
	Left     int64  `db:"lft" json:"-"`
	Right    int64  `db:"rgt" json:"-"`
	ImageURL string `json:"imageUrl" db:"image_url"`
}

type CategoryStore struct {
	bond.Store
}

func (p *Category) CollectionName() string {
	return `categories`
}

func (store CategoryStore) FindAncestors(ID int64) ([]*Category, error) {
	var categories []*Category
	err := DB.Select("cp.*").
		From("categories cc").
		Join("categories cp").
		On("cc.lft BETWEEN cp.lft AND cp.rgt").
		Where(db.And(
			db.Cond{"cc.id": ID},
			// do not return the current node as an ancestor node
			db.Cond{"cp.id": db.NotEq(ID)},
		)).
		OrderBy("id").
		All(&categories)
	return categories, err
}

func (store CategoryStore) FindDescendants(ID int64) ([]*Category, error) {
	var categories []*Category
	err := DB.Select("cc.*").
		From("categories cp").
		Join("categories cc").
		On("cc.lft BETWEEN cp.lft AND cp.rgt").
		Where(db.And(
			db.Cond{"cp.id": ID},
			// do not return the parent node when listing descs
			db.Cond{"cc.id": db.NotEq(ID)},
		)).
		OrderBy("id").
		All(&categories)
	return categories, err
}

func (store CategoryStore) FindByID(ID int64) (*Category, error) {
	return store.FindOne(db.Cond{"id": ID})
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
