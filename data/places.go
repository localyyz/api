package data

import "time"

// For now, just map places to google id
// NOTE: we would have to carry this data on our side eventually
type Place struct {
	ID int64 `db:"id,pk,omitempty" json:"id,omitempty"`
	// external google places id
	GoogleID string `db:"google_id" json:"googleId"`
	LocaleID int64  `db:"locale_id" json:"localeId"`

	Name    string `db:"name" json:"name"`
	Address string `db:"address" json:"address"`
	Phone   string `db:"phone" json:"phone"`
	Website string `db:"website" json:"website"`

	Etc *PlaceEtc `db:"etc,jsonb" json:"etc"`
	// TODO: Hours

	CreatedAt *time.Time `db:"created_at,omitempty" json:"createdAt,omitempty"`
	UpdatedAt *time.Time `db:"updated_at,omitempty" json:"updatedAt,omitempty"`
}

type PlaceWithPromo struct {
	*Place
	Promos []*Promo `json:"promos"`
}

type PlaceWithLocale struct {
	*Place
	Locale *Locale `json:"locale"`
}

type PlaceWithPost struct {
	*Place
	Posts []*PostPresenter `json:"posts"`
}
type PlaceEtc struct {
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	GoogleRating float64 `json:"google_rating"`
}

func (p *Place) CollectionName() string {
	return `places`
}
