package store

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	db "upper.io/db.v2"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/tools/blogto/locale"

	"github.com/gosimple/slug"
	"github.com/goware/geotools"
	"github.com/pkg/errors"
)

const (
	ListingURL = "http://www.blogto.com/api/v2/listings/"
	storePath  = "./data/stores/%s.json"
)

func LoadListing() error {
	files, err := ioutil.ReadDir("./data/stores")
	if err != nil {
		return err
	}

	for _, finfo := range files {
		f, err := os.Open(fmt.Sprintf("./data/stores/%s", finfo.Name()))
		if err != nil {
			return err
		}
		log.Printf("loading %s", f.Name())

		var stores []*Store
		buf, _ := ioutil.ReadAll(f)
		if err := json.Unmarshal(buf, &stores); err != nil {
			log.Fatal(err)
		}

		// locale
		for _, s := range stores {
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
			if count > 0 || err != nil {
				if err != nil {
					return err
				}
				continue
			}

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
				log.Printf("save error(%s): %v", s.Name, err)
			}
		}
	}

	return nil
}

func GetListing() error {
	for sh, locale := range locale.LocaleMap {
		log.Printf("Loading stores for %s", locale.Name)
		// check if file exists, if it does, skip

		if _, err := os.Stat(fmt.Sprintf(storePath, sh)); !os.IsNotExist(err) {
			log.Println("already loaded")
			continue
		}

		u := url.Values{
			"bundle_type":          {"large"},
			"type":                 {"5"}, // fashion, for full list, check types.go
			"limit":                {"500"},
			"offset":               {"0"},
			"ordering":             {"smart_review"},
			"default_neighborhood": {strconv.FormatInt(locale.ID, 10)},
		}
		resp, err := http.Get(fmt.Sprintf("%s?%s", ListingURL, u.Encode()))
		if err != nil {
			return errors.Wrapf(err, "store list failed for %s", locale.Name)
		}
		defer resp.Body.Close()

		var rspWrapper struct {
			Count    int      `json:"count"`
			Next     string   `json:"next"`
			Previous string   `json:"previous"`
			Results  []*Store `json:"results"`
		}
		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(&rspWrapper); err != nil {
			return errors.Wrap(err, "decode error")
		}
		log.Printf("Found %d stores", rspWrapper.Count)

		storeFile, err := os.Create(fmt.Sprintf(storePath, sh))
		if err != nil {
			return errors.Wrap(err, "create store file")
		}
		defer storeFile.Close()

		b, err := json.Marshal(rspWrapper.Results)
		if err != nil {
			return errors.Wrap(err, "encode details")
		}
		storeFile.Write(b)
	}
	return nil
}
