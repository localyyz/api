package data

import (
	"context"
	"errors"
	"strings"
	"time"

	"upper.io/db"

	"github.com/goware/geotools"
	"github.com/goware/lg"

	"googlemaps.github.io/maps"
)

var (
	mapsClient  *maps.Client
	radiusLimit = 50
	timeout     = 10 * time.Second
)

func SetupMapsClient(apiKey string) error {
	var err error
	mapsClient, err = maps.NewClient(maps.WithAPIKey(apiKey))
	return err
}

func parseAddress(address string) (parsed string) {
	parsed = address
	if address == "" {
		return
	}

	splits := strings.Split(address, ",")
	if len(splits) > 0 {
		parsed = splits[0]
	}

	return
}

func GetPlaceDetail(ctx context.Context, placeID string) (*Place, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	detailsReq := maps.PlaceDetailsRequest{
		PlaceID: placeID,
	}

	res, err := mapsClient.PlaceDetails(ctx, &detailsReq)
	if err != nil {
		return nil, err
	}

	//find locale
	//var floc bool
	//var locale *Locale
	//for _, loc := range res.AddressComponents {
	//for _, loct := range loc.Types {
	//if loct == "locality" {
	//floc = true
	//break
	//}
	//}
	//if floc {
	//locale, err := DB.Locale.FindByName(loc.ShortName)
	//if err != nil && err != db.ErrNoMoreRows {
	//return nil, err
	//}
	//if locale == nil {
	//locale = &Locale{
	//Name:        loc.ShortName,
	//Description: loc.LongName,
	//}
	//if err := DB.Locale.Save(locale); err != nil {
	//return nil, err
	//}
	//}
	//break
	//}
	//}
	place := &Place{
		GoogleID: placeID,
		Name:     res.Name,
		Address:  parseAddress(res.FormattedAddress),
		Phone:    res.FormattedPhoneNumber,
		Website:  res.Website,
	}
	//if locale != nil {
	//place.LocaleID = locale.ID
	//}

	return place, nil
}

func GetNearby(ctx context.Context, geo *geotools.Point) ([]*Place, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	latlng := geotools.LatLngFromPoint(*geo)
	lg.Warnf("nearby: lat(%v),lng(%v)", latlng.Lat, latlng.Lng)

	placeType := ctx.Value("place.type").(maps.PlaceType)
	nearbyReq := &maps.NearbySearchRequest{
		Location: &maps.LatLng{Lat: latlng.Lat, Lng: latlng.Lng},
		Radius:   75,
		Type:     placeType,
	}
	nearbyResponse, err := mapsClient.NearbySearch(ctx, nearbyReq)
	if err != nil {
		return nil, err
	}

	places := make([]*Place, len(nearbyResponse.Results))
	googleIDs := make([]string, len(places))
	for i, p := range nearbyResponse.Results {
		places[i] = &Place{
			GoogleID: p.PlaceID,
			Name:     p.Name,
			Address:  p.Vicinity,
		}
		googleIDs[i] = p.PlaceID
	}

	dbPlaceMap := map[string]int64{}
	var pl *Place
	q := DB.Place.Find(db.Cond{"google_id IN": googleIDs})
	for {
		err := q.Next(&pl)
		if err != nil {
			if err == db.ErrNoMoreRows {
				break
			}
			return nil, err
		}
		dbPlaceMap[pl.GoogleID] = pl.ID
	}

	for _, place := range places {
		plID, found := dbPlaceMap[place.GoogleID]
		if found {
			place.ID = plID
		}
	}

	return places, nil
}

func GetLocale(ctx context.Context, geo *geotools.Point) (*Locale, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	latlng := geotools.LatLngFromPoint(*geo)
	lg.Warnf("locale: lat(%v),lng(%v)", latlng.Lat, latlng.Lng)

	geocodeReq := &maps.GeocodingRequest{
		LatLng:     &maps.LatLng{Lat: latlng.Lat, Lng: latlng.Lng},
		ResultType: []string{"neighborhood"},
	}
	geocodeResponse, err := mapsClient.Geocode(ctx, geocodeReq)
	if err != nil {
		return nil, err
	}

	var locale *Locale
	for _, r := range geocodeResponse {
		ac := r.AddressComponents[0]

		locale, err = DB.Locale.FindByGoogleID(r.PlaceID)
		if err != nil {
			if err == db.ErrNoMoreRows {
				locale = &Locale{
					Name:        ac.ShortName,
					Description: ac.LongName,
					GoogleID:    r.PlaceID,
				}
				DB.Locale.Save(locale) // silently fail if needed
			}
			return nil, err
		}
		break
	}
	if locale.ID == 0 {
		return nil, errors.New("unknown locale")
	}

	return locale, nil
}
