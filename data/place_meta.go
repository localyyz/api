package data

import (
	"upper.io/bond"
	db "upper.io/db.v3"
)

type PlaceMeta struct {
	PlaceID     int64      `db:"place_id" json:"placeID"`
	Gender      *Gender    `db:"gender"`
	StyleFemale PlaceStyle `db:"style_female"`
	StyleMale   PlaceStyle `db:"style_male"`
	Pricing     string     `db:"pricing"`
}

type Gender string
type PlaceStyle string

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
