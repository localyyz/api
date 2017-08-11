package data

import (
	"time"

	"upper.io/bond"
	db "upper.io/db.v3"
)

type ShopifyCred struct {
	ID      int64 `db:"id,pk,omitempty" json:"id,omitempty"`
	PlaceID int64 `db:"place_id" json:"placeId"`

	// NOTE: because shopify is stupid
	ApiURL      string `db:"api_url" json:"api_url"`
	AccessToken string `db:"auth_access_token" json:"-"`

	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt,omitempty"`
	UpdatedAt *time.Time `db:"updated_at,omitempty" json:"updatedAt,omitempty"`
}

type ShopifyCredStore struct {
	bond.Store
}

func (s *ShopifyCred) CollectionName() string {
	return `shopify_creds`
}

func (store ShopifyCredStore) FindByPlaceID(placeID int64) (*ShopifyCred, error) {
	return store.FindOne(db.Cond{"place_id": placeID})
}

func (store ShopifyCredStore) FindOne(cond db.Cond) (*ShopifyCred, error) {
	var cred *ShopifyCred
	if err := store.Find(cond).One(&cred); err != nil {
		return nil, err
	}
	return cred, nil
}

func (store ShopifyCredStore) FindAll(cond db.Cond) ([]*ShopifyCred, error) {
	var creds []*ShopifyCred
	if err := store.Find(cond).All(&creds); err != nil {
		return nil, err
	}
	return creds, nil
}
