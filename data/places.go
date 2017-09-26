package data

import (
	"time"

	"github.com/goware/geotools"
)

type Place struct {
	ID       int64       `db:"id,pk,omitempty" json:"id,omitempty"`
	LocaleID int64       `db:"locale_id" json:"localeId"`
	Status   PlaceStatus `db:"-" json:"status"`

	Name        string `db:"name" json:"name"`
	Address     string `db:"address" json:"address"`
	Phone       string `db:"phone" json:"phone"`
	Website     string `db:"website" json:"website"`
	Description string `db:"description" json:"description"`
	ImageURL    string `db:"image_url" json:"imageUrl"`

	ShopifyID string         `db:"shopify_id,omitempty" json:"-"`
	BlogtoID  *int64         `db:"blogto_id,omitempty" json:"-"`
	Geo       geotools.Point `db:"geo" json:"-"`
	Distance  float64        `db:"distance,omitempty" json:"distance"` // calculated, not stored in db

	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt,omitempty"`
	UpdatedAt *time.Time `db:"updated_at,omitempty" json:"updatedAt,omitempty"`
}

func (p *Place) CollectionName() string {
	return `places`
}

type PlaceStatus uint32

const (
	PlaceStatusUnknown PlaceStatus = iota
	PlaceStatusInactive
	PlaceStatusActive
)
