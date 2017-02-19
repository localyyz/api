package data

import (
	"upper.io/bond"
	"upper.io/db.v3"
)

type Locale struct {
	ID          int64  `db:"id,pk,omitempty" json:"id,omitempty"`
	Name        string `db:"name" json:"name"`
	Shorthand   string `db:"shorthand" json:"shorthand"`
	Description string `db:"description" json:"description"`
}

type LocaleStore struct {
	bond.Store
}

var (
	EnabledLocales = db.Cond{"shorthand": []string{
		//"king-west",
		//"queen-west",
		//"distillery-district",
		//"kensington-market",
		//"st-lawrence-market",
		"west-queen-west",
		"yorkville",
	}}
)

func (n *Locale) CollectionName() string {
	return `locales`
}

func (store *LocaleStore) FindByName(name string) (*Locale, error) {
	return store.FindOne(db.Cond{"name": name})
}

func (store LocaleStore) FindByID(localeID int64) (*Locale, error) {
	return store.FindOne(db.Cond{"id": localeID})
}

func (store LocaleStore) FindOne(cond db.Cond) (*Locale, error) {
	var locale *Locale
	err := DB.Locale.Find(cond).One(&locale)
	if err != nil {
		return nil, err
	}
	return locale, nil
}
