package data

import (
	"time"

	"upper.io/bond"
	db "upper.io/db.v3"
)

type UserAddress struct {
	ID     int64 `db:"id,pk,omitempty" json:"id"`
	UserID int64 `db:"user_id" json:"userId"`

	FirstName    string `db:"first_name" json:"firstName"`
	LastName     string `db:"last_name" json:"lastName"`
	Address      string `db:"address" json:"address"`
	AddressOpt   string `db:"address_opt" json:"addressOpt"`
	City         string `db:"city" json:"city"`
	Country      string `db:"country" json:"country"`
	CountryCode  string `db:"country_code" json:"countryCode"`
	Province     string `db:"province" json:"province"`
	ProvinceCode string `db:"province_code" json:"provinceCode"`
	Zip          string `db:"zip" json:"zip"`

	IsShipping bool `db:"is_shipping" json:"isShipping"`
	IsBilling  bool `db:"is_billing" json:"isBilling"`

	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt,omitempty"`
	UpdatedAt *time.Time `db:"updated_at,omitempty" json:"updatedAt,omitempty"`
	DeletedAt *time.Time `db:"deleted_at,omitempty" json:"deletedAt,omitempty"`
}

type UserAddressStore struct {
	bond.Store
}

var _ interface {
	bond.HasBeforeCreate
} = &UserAddress{}

func (u *UserAddress) CollectionName() string {
	return `user_addresses`
}

func (u *UserAddress) BeforeCreate(bond.Session) error {
	u.CreatedAt = GetTimeUTCPointer()
	return nil
}

func (store UserAddressStore) FindByUserID(ID int64) ([]*UserAddress, error) {
	return store.FindAll(db.Cond{"user_id": ID})
}

func (store UserAddressStore) FindAll(cond db.Cond) ([]*UserAddress, error) {
	var addresses []*UserAddress
	if err := store.Find(cond).All(&addresses); err != nil {
		return nil, err
	}
	return addresses, nil
}
