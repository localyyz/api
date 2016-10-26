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

	"bitbucket.org/moodie-app/moodie-api/cmd/blogto/locale"

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

	var stores []*StoreDetail
	dec := json.NewDecoder(f)
	if err := dec.Decode(&stores); err != nil {
		return errors.Wrap(err, "decode error")
	}

	//for _, p := range places {
	//var dbp *data.Place
	//err = data.DB.Place.Find(
	//db.Cond{"locale_id": , "name": p["name"]},
	//).One(&dbp)
	//if err != nil {
	//if err == db.ErrNoMoreRows {
	//dbp = &data.Place{
	//Name: p["name"],
	//}
	//}
	//log.Printf("%s: %v", p["name"], err)
	//continue
	//}

	//dbp.Description = p["excerpt"].(string)
	//dbp.ImageURL = p["image"].(string)
	//dbp.Website = p["website"].(string)
	//if err := data.DB.Place.Save(dbp); err != nil {
	//log.Printf("Saving %s failed: %v\n", dbp.Name, err)
	//continue
	//}

	//lat, _ := strconv.ParseFloat(p["latitude"].(string), 64)
	//lng, _ := strconv.ParseFloat(p["longitude"].(string), 64)

	//cell, err := data.DB.Cell.FindByLatLng(lat, lng)
	//if err != nil {
	//log.Printf("place(%s) loc(%v, %v) error %s", p["name"], lat, lng, err)
	//continue
	//}

	//err = data.DB.Place.Save(&data.Place{
	//LocaleID:    cell.LocaleID,
	//Name:        strings.TrimSpace(p["name"].(string)),
	//Address:     strings.TrimSpace(p["address"].(string)),
	//Phone:       p["phone"].(string),
	//Description: p["excerpt"].(string),
	//ImageURL:    p["image"].(string),
	//Website:     p["website"].(string),
	//Geo:         *geotools.NewPointFromLatLng(lat, lng),
	//})
	//if err != nil {
	//return err
	//}
	//}

	for _, s := range stores {
		fmt.Println(s.Name)
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

	detailFile, err := os.Create(fmt.Sprintf("cmd/blogto/data/%s.detail.json", shorthand))
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
