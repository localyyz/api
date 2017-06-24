package data

import (
	"time"

	"upper.io/bond"
	"upper.io/db.v3"
)

// TODO: promo should be keyed on placeid and queried on cells
type Promo struct {
	ID        int64       `db:"id,pk,omitempty" json:"id,omitempty"`
	PlaceID   int64       `db:"place_id" json:"placeId"`
	UserID    int64       `db:"user_id" json:"userId"`
	ProductID int64       `db:"product_id" json:"productId"`
	Status    PromoStatus `db:"status" json:"status"`

	// Limits
	Limits      int64  `db:"limits" json:"limits"`
	Description string `db:"description" json:"description"`
	// external offer id refering to specific promotion
	OfferID int64    `db:"offer_id" json:"-"`
	Etc     PromoEtc `db:"etc,jsonb" json:"etc"`

	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt"`
	UpdatedAt *time.Time `db:"updated_at,omitempty" json:"updatedAt"`
	DeletedAt *time.Time `db:"deleted_at,omitempty" json:"deletedAt"`
}

type PromoStore struct {
	bond.Store
}

type PromoEtc struct {
	Price float64 `json:"prc"`
	Sku   string  `json:"sku"`
}

type (
	PromoStatus uint32
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

	p.UpdatedAt = nil
	p.CreatedAt = GetTimeUTCPointer()

	return nil
}

func (p *Promo) BeforeUpdate(bond.Session) error {
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
