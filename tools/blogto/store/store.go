package store

import (
	"strconv"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/gosimple/slug"
	"github.com/goware/geotools"
	"github.com/goware/lg"
	"github.com/pkg/errors"
	db "upper.io/db.v3"
)

type Store struct {
	Address                 string        `json:"address"`
	AppearsInBestOfLists    bool          `json:"appears_in_best_of_lists"`
	Coordinates             Coordinate    `json:"coordinates"`
	DatePublished           string        `json:"date_published"`
	DefaultNeighborhood     Neighbourhood `json:"default_neighborhood"`
	DinesafeEstablishmentID string        `json:"dinesafe_establishment_id"`
	ID                      int           `json:"id"`
	ImageURL                string        `json:"image_url"`
	Name                    string        `json:"name"`
	Phone                   string        `json:"phone"`
	Rating                  float64       `json:"rating"`
	ShareURL                string        `json:"share_url"`
	SubType                 Category      `json:"sub_type"`
	Type                    Category      `json:"type"`
	Website                 string        `json:"website"`
	IsShopify               bool          `json:"isShopify"`
}

func (s *Store) DBSave() error {
	localeSh := slug.Make(s.DefaultNeighborhood.Name)

	locale, err := data.DB.Locale.FindOne(db.Cond{"shorthand": localeSh})
	if err != nil {
		return err
	}

	count, err := data.DB.Place.Find(
		db.Cond{
			"locale_id": locale.ID,
			"name":      s.Name,
		},
	).Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	lg.Printf("inserting %s @ %s", s.Name, locale.Name)

	lat, _ := strconv.ParseFloat(s.Coordinates.Latitude, 64)
	lng, _ := strconv.ParseFloat(s.Coordinates.Longitude, 64)

	dbPlace := &data.Place{
		Name:     strings.TrimSpace(s.Name),
		Address:  strings.TrimSpace(s.Address),
		Geo:      *geotools.NewPointFromLatLng(lat, lng),
		LocaleID: locale.ID,
		ImageURL: s.ImageURL,
		Website:  s.Website,
		Phone:    s.Phone,
	}
	if err := data.DB.Place.Save(dbPlace); err != nil {
		return errors.Wrapf(err, "save error(%s): %v", s.Name)
	}

	return nil
}
