package tool

import (
	"log"
	"net/http"
	"sync"
	"time"

	"bitbucket.org/moodie-app/moodie-api/data"
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

type social struct {
	Type  socialType
	Match string

	// generated
	URL         string
	Place       *data.Place
	FBRating    *fbRating
	InstaRating *instaRating
}

type socialType uint32

func (s socialType) String() string {
	return socialTypes[s]
}

const (
	_ socialType = iota
	socialTypeFacebook
	socialTypeInstagram
)

var socialTypes = []string{"-", "facebook", "instagram"}

var socials = []social{
	{Type: socialTypeFacebook, Match: "facebook.com"},
	{Type: socialTypeInstagram, Match: "instagram.com"},
}

func fetchPlaceSocialURL(place *data.Place, socialChann chan social) error {
	//lg.Infof("fetching %s", place.Website)

	//resp, err := http.Get(place.Website)
	//if err != nil {
	//return errors.Wrapf(err, "req: place(%d)", place.ID)
	//}

	//if resp.StatusCode != 200 {
	//return fmt.Errorf("status: place(%d) with %s", place.ID, resp.Status)
	//}

	//doc, err := goquery.NewDocumentFromReader(resp.Body)
	//if err != nil {
	//return errors.Wrapf(err, "doc: place(%d)", place.ID)
	//}

	//doc.Find("a").Each(func(i int, s *goquery.Selection) {
	//// For each item found, get the band and title
	//if val, found := s.Attr("href"); found {
	//for _, sc := range socials {
	//if strings.Contains(val, sc.Match) {

	//// pass on sharer links
	//if strings.Contains(val, "share") {
	//continue
	//}

	//// do some cleaning up and make sure it's a
	//// valid url
	//u, err := url.Parse(val)
	//if err != nil {
	//log.Printf("%s: place(%d) url parse with %v", sc.Type, place.ID, err)
	//continue
	//}
	//if u.Scheme != "https" {
	//u.Scheme = "https"
	//}

	//soc := sc
	//soc.Place = place
	//soc.URL = u.String()

	//socialChann <- soc
	//}
	//}
	//}
	//})

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
	socialChann := make(chan social)

	var sg sync.WaitGroup
	for i := 0; i < 20; i++ {
		go func(placeChann chan *data.Place) {
			sg.Add(1)
			for place := range placeChann {
				//if err := fetchPlaceSocialURL(place, socialChann); err != nil {
				//log.Println(err)
				//continue
				//}

				// fetch facebook data
				//if place.FacebookURL != "" {
				//if err := fetchFacebookMeta(place, socialChann); err != nil {
				//lg.Warn(err)
				//}
				//}

				if place.InstagramURL != "" {
					rating, err := fetchInstagramMeta(place.InstagramURL)
					if err != nil {
						lg.Warnf("fetch insta error: %v", err)
						continue
					}
					socialChann <- social{
						InstaRating: rating,
						Place:       place,
						Type:        socialTypeInstagram,
					}
				}
			}
			sg.Done()
		}(placeChann)
	}

	go func(socialChann chan social) {
		for soc := range socialChann {
			//consume social
			switch soc.Type {
			case socialTypeFacebook:
				if soc.FBRating != nil {
					log.Printf("got %s rating for place(%d)\n", socialTypes[soc.Type], soc.Place.ID)
					soc.Place.Ratings.Rating = soc.FBRating.StarRating
					soc.Place.Ratings.Count = soc.FBRating.RatingCount
					soc.Place.Ratings.FBFans = soc.FBRating.FanCount
				}
				if soc.Place.FacebookURL == "" && soc.URL != "" {
					log.Printf("got %s url for place(%d)\n", socialTypes[soc.Type], soc.Place.ID)
					soc.Place.FacebookURL = soc.URL
				}
			case socialTypeInstagram:
				if rating := soc.InstaRating; rating != nil {
					log.Printf("got %s rating for place(%d): %+v \n", socialTypes[soc.Type], soc.Place.ID, rating)
					soc.Place.Ratings.InstFollowers = rating.Graphql.User.EdgeFollowedBy.Count
				}
				if soc.Place.InstagramURL == "" {
					soc.Place.InstagramURL = soc.URL
				}
			}
			if err := data.DB.Place.Save(soc.Place); err != nil {
				log.Printf("place: id %d with err %v", soc.Place.ID, err)
			}
		}
	}(socialChann)

	// producer
	var places []*data.Place
	err := data.DB.Place.Find(
		db.Cond{
			"status":        data.PlaceStatusActive,
			"created_at":    db.Gt("2018-04-15"),
			"instagram_url": db.NotEq(""),
		},
	).Limit(1).OrderBy("-id").All(&places)
	if err != nil {
		log.Println(err)
		return
	}
	for _, p := range places {
		placeChann <- p
		time.Sleep(500 * time.Millisecond)
	}
	close(placeChann)

	sg.Wait()

	render.Respond(w, r, "")
}
