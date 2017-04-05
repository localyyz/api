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

	"bitbucket.org/moodie-app/moodie-api/tools/blogto/locale"

	"github.com/pkg/errors"
)

const (
	ListingURL = "http://www.blogto.com/api/v2/listings/"
)

func LoadListing() error {
	cat := CategoryTypeFashion
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

		for _, s := range stores {
			//if err := s.DBSave(); err != nil {
			//lg.Warn(err)
			//}
			s.CheckIsShopify()
		}
	}

	return nil
}

func GetListing() error {
	cat := CategoryTypeFashion
	catStr := categoryTypes[cat]

	storeDirPath := fmt.Sprintf("./data/stores/%s", catStr)
	if _, err := os.Stat(storeDirPath); os.IsNotExist(err) {
		if err = os.Mkdir(storeDirPath, 0755); err != nil {
			log.Fatal(err)
		}
	}

	for sh, locale := range locale.LocaleMap {
		if sh != "yorkville" && sh != "west-queen-west" {
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
