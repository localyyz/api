package data

import (
	"fmt"
	"time"

	"upper.io/bond"
	db "upper.io/db.v3"
)

type ShopifyCred struct {
	ID      int64 `db:"id,pk,omitempty" json:"id,omitempty"`
	PlaceID int64 `db:"place_id" json:"placeId"`

	// NOTE: because shopify is stupid
	ApiURL      string            `db:"api_url" json:"api_url"`
	AccessToken string            `db:"auth_access_token" json:"-"`
	Status      ShopifyCredStatus `db:"status" json:"status"`

	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt,omitempty"`
	UpdatedAt *time.Time `db:"updated_at,omitempty" json:"updatedAt,omitempty"`
}

type ShopifyCredStore struct {
	bond.Store
}

type ShopifyCredStatus uint

const (
	_ ShopifyCredStatus = iota
	ShopifyCredStatusInactive
	ShopifyCredStatusActive
)

var (
	shopifyCredStatuses = []string{"unknown", "inactive", "active"}
)

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

// String returns the string value of the status.
func (s ShopifyCredStatus) String() string {
	return shopifyCredStatuses[s]
}

// MarshalText satisfies TextMarshaler
func (s ShopifyCredStatus) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalText satisfies TextUnmarshaler
func (s *ShopifyCredStatus) UnmarshalText(text []byte) error {
	enum := string(text)
	for i := 0; i < len(shopifyCredStatuses); i++ {
		if enum == shopifyCredStatuses[i] {
			*s = ShopifyCredStatus(i)
			return nil
		}
	}
	return fmt.Errorf("unknown shopify cred status %s", enum)
}
