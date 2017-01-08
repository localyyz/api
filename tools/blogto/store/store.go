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
)

func LoadListing() error {
	cat := CategoryTypeService
	catStr := categoryTypes[cat]
	storeDirPath := fmt.Sprintf("./data/stores/%s", catStr)

	files, err := ioutil.ReadDir(storeDirPath)
	if err != nil {
		return err
	}

	for _, finfo := range files {
		f, err := os.Open(fmt.Sprintf("%s/%s", storeDirPath, finfo.Name()))
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
			log.Printf("inserting %s @ %s", s.Name, locale.Name)

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
	cat := CategoryTypeService
	catStr := categoryTypes[cat]

	storeDirPath := fmt.Sprintf("./data/stores/%s", catStr)
	if _, err := os.Stat(storeDirPath); os.IsNotExist(err) {
		if err = os.Mkdir(storeDirPath, 0755); err != nil {
			log.Fatal(err)
		}
	}

	for sh, locale := range locale.LocaleMap {
		if sh != "king-west" && sh != "queen-west" {
			continue
		}
		log.Printf("Loading %s for %s", catStr, locale.Name)

		// check if file exists, if it does, skip
		filePath := fmt.Sprintf("%s/%s.json", storeDirPath, sh)
		if _, err := os.Stat(filePath); err != nil && !os.IsNotExist(err) {
			if os.IsExist(err) {
				log.Println("already loaded")
				continue
			}
			log.Fatal(err)
		}

		u := url.Values{
			"bundle_type":          {"large"},
			"type":                 {fmt.Sprintf("%d", cat)}, // for full list, check types.go
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

		storeFile, err := os.Create(filePath)
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
