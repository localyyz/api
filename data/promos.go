package data

import (
	"errors"
	"fmt"
	"time"

	"upper.io/bond"
	"upper.io/db.v3"
)

// TODO: promo should be keyed on placeid and queried on cells
type Promo struct {
	ID        int64       `db:"id,pk,omitempty" json:"id,omitempty"`
	PlaceID   int64       `db:"place_id" json:"placeId"`
	Type      PromoType   `db:"type" json:"type"`
	UserID    int64       `db:"user_id" json:"userId"`
	ProductID int64       `db:"product_id" json:"productId"`
	Status    PromoStatus `db:"status" json:"status"`

	// Limits
	Limits      int64  `db:"limits" json:"limits"`
	Description string `db:"description" json:"description"`
	// Uploaded image accompanying the promotion
	ImageUrl string `db:"image_url" json:"imageUrl"`
	// external offer id refering to specific promotion
	OfferID int64    `db:"offer_id" json:"-"`
	Etc     PromoEtc `db:"etc,jsonb" json:"etc"`

	StartAt   *time.Time `db:"start_at,omitempty" json:"startAt"`
	EndAt     *time.Time `db:"end_at,omitempty" json:"endAt"`
	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt"`
	UpdatedAt *time.Time `db:"updated_at,omitempty" json:"updatedAt"`
	DeletedAt *time.Time `db:"deleted_at,omitempty" json:"deletedAt"`
}

type PromoStore struct {
	bond.Store
}

type PromoEtc struct {
	Percent     int     `json:"pct"`
	Spend       int     `json:"spd"`
	Price       float64 `json:"prc"`
	Item        string  `json:"itm"`
	HeaderImage string  `json:"him"`
	PlaceImage  string  `json:"pim"`
	Sku         string  `json:"sku"`
}

type (
	PromoType   uint32
	PromoStatus uint32
)

const (
	_ PromoType = iota
	PromoTypePctDiscount
	PromoTypePrice
	PromoTypeMinSpend
	PromoTypeFreeItem
)

const (
	_ PromoStatus = iota
	PromoStatusDraft
	PromoStatusScheduled
	PromoStatusActive
	PromoStatusCompleted
	PromoStatusDeleted
)

var _ interface {
	bond.HasBeforeCreate
	bond.HasBeforeUpdate
} = &Promo{}

var (
	promoTypes = []string{
		"-",
		"pct_discount",
		"price",
		"min_spend",
		"free_itm",
	}
	PromoDistanceLimit = 200.0
)

func (p *Promo) CollectionName() string {
	return `promos`
}

func (p *Promo) BeforeCreate(sess bond.Session) error {
	// shared checks for updating promotion
	if err := p.BeforeUpdate(sess); err != nil {
		return err
	}

	if p.StartAt.Before(time.Now().UTC()) {
		return errors.New("start date cannot be in the past")
	}
	p.Status = PromoStatusScheduled
	p.UpdatedAt = nil
	p.CreatedAt = GetTimeUTCPointer()

	return nil
}

func (p *Promo) BeforeUpdate(bond.Session) error {
	if p.StartAt == nil {
		return errors.New("invalid start date")
	}
	if p.EndAt == nil {
		return errors.New("invalid end date")
	}
	if p.StartAt.After(*p.EndAt) {
		return errors.New("start date must be before end date")
	}
	p.UpdatedAt = GetTimeUTCPointer()

	return nil
}

func (p *Promo) CanUserClaim(userID int64) (bool, error) {
	count, err := DB.Claim.Find(db.Cond{"user_id": userID, "promo_id": p.ID}).Count()
	if err != nil {
		return false, err
	}
	return (count > 0), nil
}

func (store PromoStore) FindByPlaceID(placeID int64) (*Promo, error) {
	return store.FindOne(db.Cond{"place_id": placeID})
}

func (store PromoStore) FindByID(ID int64) (*Promo, error) {
	return store.FindOne(db.Cond{"id": ID})
}

func (store PromoStore) FindByOfferID(offerID int64) (*Promo, error) {
	return store.FindOne(db.Cond{"offer_id": offerID})
}

func (store PromoStore) FindOne(cond db.Cond) (*Promo, error) {
	var promo *Promo
	if err := store.Find(cond).One(&promo); err != nil {
		return nil, err
	}
	return promo, nil
}

func (store PromoStore) FindAll(cond db.Cond) ([]*Promo, error) {
	var promos []*Promo
	if err := store.Find(cond).All(&promos); err != nil {
		return nil, err
	}
	return promos, nil
}

// String returns the string value of the status.
func (pt PromoType) String() string {
	return promoTypes[pt]
}

// MarshalText satisfies TextMarshaler
func (pt PromoType) MarshalText() ([]byte, error) {
	return []byte(pt.String()), nil
}

// UnmarshalText satisfies TextUnmarshaler
func (pt *PromoType) UnmarshalText(text []byte) error {
	enum := string(text)
	for i := 0; i < len(promoTypes); i++ {
		if enum == promoTypes[i] {
			*pt = PromoType(i)
			return nil
		}
	}
	return fmt.Errorf("unknown promotion type %s", enum)
}
