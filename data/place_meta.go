package data

import (
	"github.com/pkg/errors"
	"upper.io/bond"
	db "upper.io/db.v3"
)

type PlaceMeta struct {
	PlaceID     int64       `db:"place_id" json:"placeID"`
	Gender      *Gender     `db:"gender,omitempty"`
	StyleFemale *PlaceStyle `db:"style_female,omitempty"`
	StyleMale   *PlaceStyle `db:"style_male,omitempty"`
	Pricing     string      `db:"pricing"`
}

type Gender string
type PlaceStyle string

var ErrNoPreference = errors.New("no user preference")

const (
	GenderMale   = "male"
	GenderFemale = "female"
)

type PlaceMetaStore struct {
	bond.Store
}

func (b *PlaceMeta) CollectionName() string {
	return `place_meta`
}

func (store *PlaceMetaStore) FindByPlaceID(ID int64) (*PlaceMeta, error) {
	var placeMeta *PlaceMeta
	err := store.Find(db.Cond{"place_id": ID}).One(&placeMeta)
	if err != nil {
		return nil, err
	}
	return placeMeta, nil
}

func (store *PlaceMetaStore) GetPlacesFromPreference(prf *UserPreference) ([]int64, error) {
	if prf == nil {
		return nil, ErrNoPreference
	}

	styleCol := "style_female"
	if prf.Gender[0] == "man" {
		styleCol = "style_male"
	}

	var meta []PlaceMeta
	err := store.
		Find(
			db.And(
				db.Or(
					db.Cond{"gender": prf.Gender},
					db.Cond{"gender": db.IsNull()},
				),
				db.Cond{
					"pricing": prf.Pricings,
					styleCol:  prf.Styles,
				},
			),
		).
		All(&meta)
	if err != nil {
		return nil, err
	}

	var placeIDs []int64
	for _, p := range meta {
		placeIDs = append(placeIDs, p.PlaceID)
	}

	return placeIDs, nil
}

func (store *PlaceMetaStore) GetStyles() ([]string, error) {
	rows, err := DB.Select(db.Raw(
		"unnest(enum_range(null::place_style))",
	)).Query()
	if err != nil {
		return nil, err
	}

	var styles []string
	for {
		if !rows.Next() {
			break
		}
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		styles = append(styles, s)
	}

	return styles, nil
}
