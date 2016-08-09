package data

import (
	"upper.io/bond"
	"upper.io/db"
)

type Locale struct {
	ID          int64  `db:"id,pk,omitempty" json:"id,omitempty"`
	Name        string `db:"name" json:"name"`
	Description string `db:"description" json:"description"`
	GoogleID    string `db:"google_id" json:"googleId"`
}

type LocaleStore struct {
	bond.Store
}

func (n *Locale) CollectionName() string {
	return `locales`
}

func (store *LocaleStore) FindByName(name string) (*Locale, error) {
	return store.FindOne(db.Cond{"name": name})
}

func (store LocaleStore) FindByGoogleID(googleID string) (*Locale, error) {
	return store.FindOne(db.Cond{"google_id": googleID})
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
