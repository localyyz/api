package store

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	db "upper.io/db.v2"

	"bitbucket.org/moodie-app/moodie-api/cmd/blogto/locale"
	"bitbucket.org/moodie-app/moodie-api/data"

	"github.com/goware/geotools"
	"github.com/goware/lg"
	"github.com/pkg/errors"
)

const (
	ApiURL     = "http://www.blogto.com"
	ListingURL = "http://api.blogto.com/listings.json"
)

var (
	listingRgx = regexp.MustCompile("(?:var iListingID = )(\\d*)")
)

func LoadListing(shorthand string) error {
	f, err := os.Open(fmt.Sprintf("cmd/blogto/data/%s.json", shorthand))
	if err != nil {
		return errors.Wrap(err, "file error, try list first.")
	}

	locale, err := data.DB.Locale.FindOne(db.Cond{"shorthand": shorthand})
	if err != nil {
		return err
	}

	var stores []*StoreDetail
	dec := json.NewDecoder(f)
	if err := dec.Decode(&stores); err != nil {
		return errors.Wrap(err, "decode error")
	}

	for _, s := range stores {
		blogtoID, _ := strconv.ParseInt(s.ID, 10, 64)
		dbp, err := data.DB.Place.FindOne(
			db.Cond{
				"locale_id": locale.ID,
				"blogto_id": blogtoID,
			},
		)
		if err != nil {
			if err != db.ErrNoMoreRows {
				log.Printf("find error(%s): %v", s.Name, err)
				continue
			}
			lat, _ := strconv.ParseFloat(s.Latitude, 64)
			lng, _ := strconv.ParseFloat(s.Longitude, 64)
			dbp = &data.Place{
				Name:     strings.TrimSpace(s.Name),
				Address:  strings.TrimSpace(s.Address),
				Geo:      *geotools.NewPointFromLatLng(lat, lng),
				LocaleID: locale.ID,
			}
		}
		dbp.Description = s.Excerpt
		dbp.ImageURL = s.Image
		dbp.Website = s.Website
		dbp.Phone = s.Phone
		dbp.BlogtoID = &blogtoID

		// categories
		cat := strings.ToLower(s.Category)
		if err := dbp.Category.UnmarshalText([]byte(cat)); err != nil {
			// see if it's men / women
			if strings.Contains(cat, "clothing") {
				dbp.Category = data.CategoryClothing
			} else if strings.Contains(cat, "vintage") {
				dbp.Category = data.CategoryVintage
			}
		}
		lg.Printf("name(%s) cat(%s)", dbp.Name, dbp.Category)

		// gender
		if strings.Contains(cat, "men") {
			dbp.Gender = data.PlaceGenderMale
		} else if strings.Contains(cat, "women") {
			dbp.Gender = data.PlaceGenderFemale
		}

		if err := data.DB.Place.Save(dbp); err != nil {
			log.Printf("save error(%s): %v", s.Name, err)
		}
	}

	return nil
}

func GetListing(shorthand string) error {
	locale, ok := locale.LocaleMap[shorthand]
	if !ok {
		return errors.Errorf("unknown locale %s", shorthand)
	}
	u := url.Values{
		"cmd":           {"map-search"},
		"type":          {"5"}, // fashion
		"mode":          {"searchresults"},
		"neighbourhood": {strconv.FormatInt(locale.ID, 10)},
	}
	resp, err := http.PostForm(fmt.Sprintf("%s/fashion", ApiURL), u)
	if err != nil {
		return errors.Wrap(err, "http listing failure")
	}
	defer resp.Body.Close()

	var stores []*Store
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&stores); err != nil {
		return errors.Wrap(err, "decode error")
	}
	log.Printf("Found %d stores", len(stores))

	for _, p := range stores {
		resp, _ := http.Get(fmt.Sprintf("%s/%s", ApiURL, p.URLPath))
		html, _ := ioutil.ReadAll(resp.Body)

		if m := listingRgx.FindSubmatch(html); len(m) > 1 {
			lID, _ := strconv.Atoi(string(m[1]))
			p.ListingID = lID
		}
		resp.Body.Close()
	}

	detailFile, err := os.Create(fmt.Sprintf("cmd/blogto/data/%s.json", shorthand))
	if err != nil {
		return errors.Wrap(err, "detail file")
	}
	defer detailFile.Close()

	storeDetails := make([]*StoreDetail, len(stores))
	for i, s := range stores {
		resp, _ := http.Get(fmt.Sprintf("%s/%d?k=d4UKy3LrxPXA", ListingURL, s.ListingID))
		log.Printf("Store(%d): %s", s.ListingID, s.Name)

		var detail *StoreDetail
		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(&detail); err != nil {
			log.Printf("decode error: %s", err)
			continue
		}
		storeDetails[i] = detail
	}
	b, err := json.Marshal(storeDetails)
	if err != nil {
		return errors.Wrap(err, "encode details")
	}
	detailFile.Write(b)

	return nil
}
