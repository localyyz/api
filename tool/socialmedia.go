package tool

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-chi/render"
	"github.com/pkg/errors"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
	//"fmt"
	//"log"
	//"net/http"
	//"net/url"
	//"strings"
	//"sync"
	//"time"
	//"bitbucket.org/moodie-app/moodie-api/data"
	//"github.com/PuerkitoBio/goquery"
	//"github.com/go-chi/render"
	//fb "github.com/huandu/facebook"
	//"github.com/pkg/errors"
	//"github.com/pressly/lg"
	//db "upper.io/db.v3"
)

type metatyper interface {
	GetURL() string
	GetType() metaType
	GetPlace() *data.Place
}

type social struct {
	Type  metaType
	Match string

	// generated
	URL         string
	Place       *data.Place
	FBRating    *fbRating
	InstaRating *instaRating
}

type metaType uint32

func (s metaType) String() string {
	return metaTypes[s]
}

func (u social) GetPlace() *data.Place {
	return u.Place
}

func (u social) GetType() metaType {
	return u.Type
}

func (u social) GetURL() string {
	return u.URL
}

const (
	_ metaType = iota
	socialTypeFacebook
	socialTypeInstagram
	policyTypeReturns
	policyTypeShipping
)

var metaTypes = []string{"-", "facebook", "instagram", "return", "shipping"}

var socials = []social{
	{Type: socialTypeFacebook, Match: "facebook.com"},
	{Type: socialTypeInstagram, Match: "instagram.com"},
}

func fetchURL(place *data.Place, chnn chan metatyper) error {
	lg.Infof("fetching %s", place.Website)

	resp, err := http.Get(place.Website)
	if err != nil {
		return errors.Wrapf(err, "req: place(%d)", place.ID)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("status: place(%d) with %s", place.ID, resp.Status)
	}

	//reader, _ := os.Open("tmp/daytonboots.html")
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return errors.Wrapf(err, "doc: place(%d)", place.ID)
	}

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the links
		if val, found := s.Attr("href"); found {
			for _, sc := range socials {
				if strings.Contains(val, sc.Match) {

					// pass on sharer links
					if strings.Contains(val, "share") {
						continue
					}

					// do some cleaning up and make sure it's a valid url
					u, err := url.Parse(val)
					if err != nil {
						log.Printf("%s: place(%d) url parse with %v", sc.Type, place.ID, err)
						continue
					}
					if u.Scheme != "https" {
						u.Scheme = "https"
					}
					soc := sc
					soc.Place = place
					soc.URL = u.String()

					chnn <- metatyper(soc)
				}
			}
		}
	})

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
	})

	return nil
}

type fbRating struct {
	StarRating  float32 `json:"overall_star_rating"`
	RatingCount int64   `json:"rating_count"`
	FanCount    int64   `json:"fan_count"`
	ID          string  `json:"id"`
}

func fetchFacebookMeta(place *data.Place, socialChann chan social) error {
	//u, err := url.Parse(place.FacebookURL)
	//if err != nil {
	//return errors.Wrap(err, "facebook url parse")
	//}

	//fbID := strings.Trim(u.Path, "/")
	//params := fb.Params{
	//"fields":       "overall_star_rating,rating_count,fan_count",
	//"access_token": "EAACEdEose0cBAAMeN3ZAkcajfF8LBpSSMUEmNWKZBLKBF9nkZC2f5bk3YVrDTBBZBlPCML2CfyPVjZBoAsiA5ZAB4tVUhHtdOMappbA5hT15jdLOVcpzOH1PwNfz3k8zMmuxs0O41ZBTTvcNxrlfeQfKaaTx9Yos9C9XSPrmlJkxh8UJnRCrNMQcfNLgtBby6MZD",
	//}
	//resp, err := fb.Get(fbID, params)
	//if err != nil {
	//return errors.Wrapf(err, "facebook %s get", fbID)
	//}

	//rating := &fbRating{}
	//if err := resp.Decode(rating); err != nil {
	//return errors.Wrapf(err, "facebook %s decode", fbID)
	//}

	//socialChann <- social{FBRating: rating, Place: place, Type: socialTypeFacebook}

	return nil
}

type instaRating struct {
	Graphql struct {
		User struct {
			EdgeFollowedBy struct {
				Count int64 `json:"count"`
			} `json:"edge_followed_by"`
		} `json:"user"`
	} `json:"graphql"`
}

func fetchInstagramMeta(URL string) (*instaRating, error) {
	resp, err := http.Get(URL)
	if err != nil {
		return nil, errors.Wrapf(err, "insta %s get", URL)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Wrapf(err, "insta %s get got status %s", URL, resp.Status)
	}

	//rating := &instaRating{}

	return nil, nil
}

func GetSocialMedia(w http.ResponseWriter, r *http.Request) {
	placeChann := make(chan *data.Place, 20)
	chann := make(chan metatyper)

	var sg sync.WaitGroup
	for i := 0; i < 20; i++ {
		go func(placeChann chan *data.Place) {
			sg.Add(1)
			for place := range placeChann {
				if err := fetchURL(place, chann); err != nil {
					log.Println(err)
					continue
				}

				//if place.InstagramURL != "" {
				//rating, err := fetchInstagramMeta(place.InstagramURL)
				//if err != nil {
				//lg.Warnf("fetch insta error: %v", err)
				//continue
				//}
				//socialChann <- social{
				//InstaRating: rating,
				//Place: place,
				//Type:  socialTypeInstagram,
				//}
				//}
			}
			sg.Done()
		}(placeChann)
	}

	go func(chnn chan metatyper) {
		for soc := range chnn {
			//consume social
			place := soc.GetPlace()
			log.Printf("got %s for %s: %s", soc.GetType(), place.Name, soc.GetURL())
			switch soc.GetType() {
			case socialTypeFacebook:
				//if soc.FBRating != nil {
				//log.Printf("got %s rating for place(%d)\n", socialTypes[soc.Type], soc.Place.ID)
				//soc.Place.Ratings.Rating = soc.FBRating.StarRating
				//soc.Place.Ratings.Count = soc.FBRating.RatingCount
				//soc.Place.Ratings.FBFans = soc.FBRating.FanCount
				//}
				if place.FacebookURL == "" && soc.GetURL() != "" {
					log.Printf("got fb url for place(%d)", place.ID)
					place.FacebookURL = soc.GetURL()
				}
			case socialTypeInstagram:
				//if rating := soc.InstaRating; rating != nil {
				//log.Printf("got %s rating for place(%d): %+v \n", socialTypes[soc.Type], soc.Place.ID, rating)
				//soc.Place.Ratings.InstFollowers = rating.Graphql.User.EdgeFollowedBy.Count
				//}
				if place.InstagramURL == "" {
					place.InstagramURL = soc.GetURL()
				}
			}

			if err := data.DB.Place.Save(place); err != nil {
				log.Printf("place: with err %v", err)
			}
		}
	}(chann)

	// producer
	var places []*data.Place
	err := data.DB.Place.Find(
		db.Cond{
			"status": data.PlaceStatusActive,
			db.Raw("shipping_policy->>'desc'"): db.Eq(""),
		},
	).All(&places)
	if err != nil {
		log.Println(err)
		return
	}
	for _, p := range places {
		placeChann <- p
		time.Sleep(5 * time.Second)
	}
	close(placeChann)

	sg.Wait()

	render.Respond(w, r, "")
}
