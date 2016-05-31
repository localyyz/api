package data

import (
	"time"

	"upper.io/bond"
	"upper.io/db"
)

type Account struct {
	ID        int64  `db:"id,pk,omitempty" json:"id" facebook:"-"`
	Username  string `db:"username" json:"username" facebook:"id,required"`
	Email     string `db:"email" json:"email" facebook:"email"`
	Name      string `db:"name" json:"name" facebook:"name"`
	AvatarURL string `db:"avatar_url" json:"avatar_url"`

	AccessToken string     `db:"access_token" json:"access_token"`
	Network     string     `db:"network" json:"network"`
	ExpiresAt   *time.Time `db:"expires_at" json:"expires_at"`

	CreatedAt *time.Time `db:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt *time.Time `db:"updated_at,omitempty" json:"updated_at,omitempty"`
	DeletedAt *time.Time `db:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

type AccountStore struct {
	bond.Store
}

func (u *Account) CollectionName() string {
	return `accounts`
}

func (s AccountStore) FindByUsername(username string) (*Account, error) {
	return s.FindOne(db.Cond{"username": username})
}

func (s AccountStore) FindOne(cond db.Cond) (*Account, error) {
	var a *Account
	if err := s.Find(cond).One(&a); err != nil {
		return nil, err
	}
	return a, nil
}
