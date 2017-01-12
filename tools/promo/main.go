package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/goware/lg"
	"github.com/pkg/errors"

	"bitbucket.org/moodie-app/moodie-api/data"
)

const timeForm = `2006-01-02 15:04:05`

var NonAlpha = regexp.MustCompile(`\W+`)
var RemoveBracket = regexp.MustCompile(`\s\((.*?)\)`)

func ReadCSV(input io.Reader) ([]*data.Promo, error) {
	reader := csv.NewReader(input)
	reader.TrimLeadingSpace = true

	var promos []*data.Promo
	// read one line for the header data
	csvHeader, err := reader.Read()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read csv header")
	}

	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break // done
			}
			return nil, errors.Wrap(err, "failed to read csv data")
		}

		p, err := parseRecord(record, csvHeader)
		if err != nil {
			lg.Warn(err)
			continue // silently continue. TODO: count error
		}
		promos = append(promos, p)
	}

	return promos, nil
}

func parseRecord(record []string, header []string) (*data.Promo, error) {
	var (
		p       = new(data.Promo)
		err     error
		startAt time.Time
		endAt   time.Time
	)
	for i, field := range record {
		switch header[i] {
		case "place_id":
			placeID, _ := strconv.ParseInt(field, 10, 64)
			place, err := data.DB.Place.FindByID(placeID)
			if err != nil {
				lg.Warnf("unknown place(%d)", placeID)
				continue
			}
			p.PlaceID = place.ID

			sp := NonAlpha.ReplaceAllString(RemoveBracket.ReplaceAllString(strings.ToLower(place.Name), ""), "_")
			p.Etc.PlaceImage = fmt.Sprintf("https://moodie.s3.amazonaws.com/%s.jpg", sp)
		case "description":
			p.Description = field
		case "promotion_type":
			promoType, _ := strconv.Atoi(field)
			p.Type = data.PromoType(promoType)
		case "promotion_pct":
			p.Etc.Percent, _ = strconv.Atoi(field)
		case "promotion_spend":
			p.Etc.Spend, _ = strconv.Atoi(field)
		case "promotion_item":
			p.Etc.Item = field
		case "limit":
			p.Limits, _ = strconv.ParseInt(field, 10, 64)
		case "duration":
			dur, err := time.ParseDuration(field)
			if err != nil {
				lg.Warnf("invalid duration %+v, err: %v", record, err)
				continue
			}
			p.Duration = int64(dur)
		case "start_at":
			startAt, err = time.Parse(timeForm, field)
			if err != nil {
				lg.Warnf("invalid start %+v, err: %v", field, err)
				continue
			}
		case "end_at":
			endAt, err = time.Parse(timeForm, field)
			if err != nil {
				lg.Warnf("invalid start %+v, err: %v", field, err)
				continue
			}
		default:
			continue // continue next record field
		}
	}

	// always use now as start time
	p.StartAt = data.GetTimeUTCPointer()
	d := endAt.Sub(startAt)
	e := p.StartAt.Add(d)
	p.EndAt = &e

	return p, nil
}

func LoadPromoCSV() {
	// load csv
	f, err := os.Open("promos.csv")
	if err != nil {
		log.Fatal(err)
	}

	promos, err := ReadCSV(f)
	if err != nil {
		log.Fatal(err)
	}

	for i, p := range promos {
		p.Etc.HeaderImage = fmt.Sprintf("https://moodie.s3.amazonaws.com/promo/v1/ref%d.jpg", (i + 1))
		if err := data.DB.Promo.Save(p); err != nil {
			log.Printf("saving failed %+v, error: %v", p, err)
		}
	}
}

func main() {
	fmt.Println("starting promotion loader")

	conf := &data.DBConf{
		Database:        "localyyz",
		Hosts:           []string{"localhost:5432"},
		Username:        "localyyz",
		ApplicationName: "promo loader",
	}
	if _, err := data.NewDBSession(conf); err != nil {
		log.Fatalf("db err: %s. Check ssh tunnel.", err)
	}

}
