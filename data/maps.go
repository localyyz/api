package data

import (
	"context"

	"github.com/goware/geotools"
	"github.com/goware/lg"

	"googlemaps.github.io/maps"
)

var (
	mapsClient  *maps.Client
	radiusLimit = 50
)

func SetupMapsClient(apiKey string) error {
	var err error
	mapsClient, err = maps.NewClient(maps.WithAPIKey(apiKey))
	return err
}

func GetNearby(ctx context.Context, geo *geotools.Point) ([]*Place, error) {
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
	for i, p := range nearbyResponse.Results {
		places[i] = &Place{
			GoogleID: p.PlaceID,
			Name:     p.Name,
			Address:  p.Vicinity,
		}
	}
	return places, nil
}

func GetLocale(ctx context.Context, geo *geotools.Point) (*Locale, error) {
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
		locale = &Locale{
			Name:        ac.ShortName,
			Description: ac.LongName,
			GoogleID:    r.PlaceID,
		}
		break
	}
	DB.Locale.Save(locale) // silently fail if needed

	return locale, nil
}
