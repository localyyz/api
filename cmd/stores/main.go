package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
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
}

const (
	ApiURL = "http://www.blogto.com/fashion"
)

var (
	flags = flag.NewFlagSet("locale", flag.ExitOnError)
	//localeID = flags.Int("id", 0, "Locale id to load")
)

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

	f, err := os.Open("cmd/stores/data/kingwest.json")
	if err != nil {
		log.Fatal(err, "file")
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err, "io")
	}

	var places []Place
	if err := json.Unmarshal(b, &places); err != nil {
		log.Fatal(err, "json")
	}

	for _, p := range places {
		lat, _ := strconv.ParseFloat(p.Latitude, 64)
		lng, _ := strconv.ParseFloat(p.Longitude, 64)

		cell, err := data.DB.Cell.FindByLatLng(lat, lng)
		if err != nil {
			log.Printf("place(%s) loc(%v, %v) error %s", p.Name, lat, lng, err)
			continue
		}

		err = data.DB.Place.Save(&data.Place{
			LocaleID: cell.LocaleID,
			Name:     strings.TrimSpace(p.Name),
			Address:  strings.TrimSpace(p.Address),
			Phone:    p.Phone,
			Geo:      *geotools.NewPointFromLatLng(lat, lng),
		})
		if err != nil {
			log.Println(err)
		}
	}
}
