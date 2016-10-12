package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/goware/geotools"

	"bitbucket.org/moodie-app/moodie-api/data"
)

type Place struct {
	Address   string `json:"Address"`
	Latitude  string `json:"Latitude"`
	Longitude string `json:"Longitude"`
	Name      string `json:"Name"`
	Phone     string `json:"Phone"`
	URLPath   string `json:"URL"`
	ListingID int    `json:"ListingID"`
}

const (
	ApiURL     = "http://www.blogto.com"
	ListingURL = "http://api.blogto.com/listings.json"
)

var (
	flags      = flag.NewFlagSet("locale", flag.ExitOnError)
	listingRgx = regexp.MustCompile("(?:var iListingID = )(\\d*)")
	//localeID = flags.Int("id", 0, "Locale id to load")
)

func writeListings(locale string) error {
	f, err := os.Open(fmt.Sprintf("data/%s.detail.json", locale))
	if err != nil {
		return err
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	var places []map[string]interface{}
	if err := json.Unmarshal(b, &places); err != nil {
		return err
	}

	for _, p := range places {
		lat, _ := strconv.ParseFloat(p["latitude"].(string), 64)
		lng, _ := strconv.ParseFloat(p["longitude"].(string), 64)

		cell, err := data.DB.Cell.FindByLatLng(lat, lng)
		if err != nil {
			log.Printf("place(%s) loc(%v, %v) error %s", p["name"], lat, lng, err)
			continue
		}

		err = data.DB.Place.Save(&data.Place{
			LocaleID:    cell.LocaleID,
			Name:        strings.TrimSpace(p["name"].(string)),
			Address:     strings.TrimSpace(p["address"].(string)),
			Phone:       p["phone"].(string),
			Description: p["excerpt"].(string),
			ImageURL:    p["image"].(string),
			Website:     p["website"].(string),
			Geo:         *geotools.NewPointFromLatLng(lat, lng),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func getListing(locale string) error {
	f, _ := os.Open(fmt.Sprintf("data/%s.json", locale))

	var places []*Place
	dec := json.NewDecoder(f)
	if err := dec.Decode(&places); err != nil {
		return err
	}

	for _, p := range places {
		resp, _ := http.Get(fmt.Sprintf("%s/%s", ApiURL, p.URLPath))
		html, _ := ioutil.ReadAll(resp.Body)

		if m := listingRgx.FindSubmatch(html); len(m) > 1 {
			lID, _ := strconv.Atoi(string(m[1]))
			p.ListingID = lID
		}

		resp.Body.Close()
	}
	f.Close()

	detail, _ := os.Create(fmt.Sprintf("data/%s.detail.json", locale))
	for _, p := range places {
		resp, _ := http.Get(fmt.Sprintf("%s/%d?k=d4UKy3LrxPXA", ListingURL, p.ListingID))
		json, _ := ioutil.ReadAll(resp.Body)
		detail.Write(json)
	}
	detail.Close()

	return nil
}

// curl -XPOST 'http://www.blogto.com/fashion' --data 'cmd=map-search&type=5&mode=searchresults&neighbourhood=29'
//u := url.Values{
//"cmd":           {"map-search"},
//"mode":          {"searchresults"},
//"neighbourhood": {strconv.Itoa(*localeID)},
//}
//resp, err := http.PostForm(ApiURL, u)
//if err != nil {
//log.Fatal(err)
//}
//defer resp.Body.Close()

func main() {
	flags.Parse(os.Args[1:])

	conf := data.DBConf{
		Database: "moodie",
		Hosts:    []string{":5432"},
		Username: "moodie",
	}
	if err := data.NewDBSession(conf); err != nil {
		log.Fatal(err, "db")
	}

	//fmt.Println(writeListings("kingwest"))
}
