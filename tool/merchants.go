package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
)

func LoadMetadata() {
	type data struct {
		ID                    int64  `json:"id"`
		FacebookURL           string `json:"facebookURL"`
		InstagramURL          string `json:"instagramURL"`
		ReturnPolicyURL       string `json:"returnPolicyURL"`
		ReturnPolicySummary   string `json:"returnPolicySummary"`
		ShippingPolicyURL     string `json:"shippingPolicyURL"`
		ShippingPolicySummary string `json:"shippingPolicySummary"`
		Ratings               string `json:"ratings"`
	}

	f, err := os.Open("./tmp/test.json")
	if err != nil {
		log.Fatal(err)
	}

	var dd []data
	json.NewDecoder(f).Decode(&dd)

	for _, data := range dd {
		ratings := "{}"
		if data.Ratings != "-" {
			rr := strings.Split(data.Ratings, ",")

			rating, _ := strconv.ParseFloat(strings.TrimSpace(rr[0]), 32)
			count, _ := strconv.ParseInt(strings.TrimSpace(rr[1]), 10, 64)
			ratings = fmt.Sprintf(`{"rating":%.2f,"count":%d}`, rating, count)
		}

		fmt.Printf(`
		UPDATE places
		   SET fb_url = '%s',
		       instagram_url = '%s',
		       shipping_policy = '{"url":"%s","desc":"%s"}',
		       return_policy = '{"url":"%s","desc":"%s"}',
		       ratings = '%s'
		WHERE id = %d;
		`,
			data.FacebookURL,
			data.InstagramURL,
			data.ShippingPolicyURL,
			data.ShippingPolicySummary,
			data.ReturnPolicyURL,
			data.ReturnPolicySummary,
			ratings,
			data.ID,
		)

	}
}

func ListPermissions(w http.ResponseWriter, r *http.Request) {
	var creds []*data.ShopifyCred

	data.DB.Select(db.Raw("c.*")).
		From("shopify_creds c").
		LeftJoin("places p").On("p.id = c.place_id").
		Where(db.Cond{"p.status": data.PlaceStatusActive}).
		All(&creds)

	type countCompare struct {
		Theirs  int   `json:"theirs"`
		Ours    int   `json:"ours"`
		PlaceID int64 `json:"placeId"`
	}
	var result []countCompare

	for _, cred := range creds {
		client := shopify.NewClient(nil, cred.AccessToken)
		client.BaseURL, _ = url.Parse(cred.ApiURL)
		_, resp, err := client.PriceRule.List(r.Context(), nil)
		if resp.StatusCode != http.StatusOK || err != nil {
			log.Printf("place %d: %+v", cred.PlaceID, err)
			continue
		}
	}

	render.Respond(w, r, result)

}

type priceRule struct {
	PlaceID                        int64  `json:"merchantId"`
	PlaceName                      string `json:"merchant"`
	PlaceWebsite                   string `json:"website"`
	Title                          string `json:"title"`
	ValueType                      string `json:"value_type"`
	Value                          string `json:"value"`
	TargetType                     string `json:"target_type"`
	PrerequisiteShippingPriceRange struct {
		Lte string `json:"less_than_or_equal_to"`
	} `json:"prerequisite_shipping_price_range"`
	PrerequisiteQuantityRange struct {
		Gte string `json:"greater_than_or_equal_to"`
	} `json:"prerequisite_quantity_range"`
	StartsAt time.Time  `json:"starts_at"`
	EndsAt   *time.Time `json:"ends_at"`
}

type placeRules struct {
	Place *data.Place
	Rules []*shopify.PriceRule
}

func ListPriceRules(w http.ResponseWriter, r *http.Request) {
	credChann := make(chan *data.ShopifyCred, 20)
	placeRulechann := make(chan placeRules)

	var sg sync.WaitGroup
	for i := 0; i < 20; i++ {
		go func(credChann chan *data.ShopifyCred) {
			sg.Add(1)
			for cred := range credChann {
				place, err := data.DB.Place.FindByID(cred.PlaceID)
				if err != nil {
					log.Printf("place fetch error %+v", err)
					continue
				}

				client := shopify.NewClient(nil, cred.AccessToken)
				client.BaseURL, _ = url.Parse(cred.ApiURL)
				priceRules, _, _ := client.PriceRule.List(
					r.Context(),
					&shopify.PriceRuleParam{
						EndsAtMin: data.GetTimeUTCPointer(),
					},
				)
				placeRulechann <- placeRules{
					Place: place,
					Rules: priceRules,
				}
				fmt.Printf(".")
			}
			sg.Done()
		}(credChann)
	}

	var creds []*data.ShopifyCred
	data.DB.Select(db.Raw("c.*")).
		From("shopify_creds c").
		LeftJoin("places p").On("p.id = c.place_id").
		Where(db.Cond{"p.status": data.PlaceStatusActive}).
		OrderBy("id DESC").
		All(&creds)

	var newRules []priceRule
	go func() {
		for prules := range placeRulechann {
			for _, r := range prules.Rules {
				newRules = append(newRules, priceRule{
					PlaceID:      prules.Place.ID,
					PlaceName:    prules.Place.Name,
					PlaceWebsite: prules.Place.Website,

					Title:      r.Title,
					TargetType: r.TargetType,
					ValueType:  r.ValueType,
					Value:      r.Value,
					PrerequisiteQuantityRange:      r.PrerequisiteQuantityRange,
					PrerequisiteShippingPriceRange: r.PrerequisiteShippingPriceRange,
					StartsAt:                       r.StartsAt,
					EndsAt:                         r.EndsAt,
				})
			}
		}
	}()

	for _, cred := range creds {
		credChann <- cred
	}
	close(credChann)

	// wait for all the producers to finish
	sg.Wait()

	render.Respond(w, r, newRules)
}

type activePlace struct {
	Name         string           `json:"name"`
	Website      string           `json:"website"`
	Email        string           `json:"email"`
	Gender       data.PlaceGender `json:"gender"`
	Country      string           `json:"country"`
	Plan         string           `json:"plan"`
	Phone        string           `json:"phone"`
	Status       data.PlaceStatus `json:"status"`
	ProductCount uint64           `json:"productCount"`

	PriceRange    data.MerchantApprovalPriceRange `json:"priceRange,omitempty"`
	Category      data.MerchantApprovalCategory   `json:"category,omitempty"`
	Collection    data.MerchantApprovalCollection `json:"collection,omitempty"`
	PriceRules    []shopify.PriceRule             `json:"price_rules,omitempty"`
	DateOnboarded time.Time                       `json:"dateOnboarded"`

	FacebookURL  string           `json:"fb_url"`
	InstagramURL string           `json:"insta_url"`
	Ratings      data.PlaceRating `json:"rating"`
}

func ListActive(w http.ResponseWriter, r *http.Request) {
	credChann := make(chan *data.ShopifyCred, 20)
	activeChann := make(chan activePlace)

	var sg sync.WaitGroup
	for i := 0; i < 20; i++ {
		go func(credChann chan *data.ShopifyCred) {
			sg.Add(1)
			for cred := range credChann {
				place, err := data.DB.Place.FindByID(cred.PlaceID)
				if err != nil {
					log.Printf("place fetch error %+v", err)
					continue
				}
				a := &activePlace{
					Name:          place.Name,
					Website:       place.Website,
					Gender:        place.Gender,
					Status:        place.Status,
					DateOnboarded: *place.CreatedAt,
					Plan:          place.Plan,
					Ratings:       place.Ratings,
					FacebookURL:   place.FacebookURL,
					InstagramURL:  place.InstagramURL,
				}
				if approval, _ := data.DB.MerchantApproval.FindByPlaceID(cred.PlaceID); approval != nil {
					a.PriceRange = approval.PriceRange
					a.Category = approval.Category
					a.Collection = approval.Collection
				}
				if place.Status == data.PlaceStatusActive {
					client := shopify.NewClient(nil, cred.AccessToken)
					client.BaseURL, _ = url.Parse(cred.ApiURL)
					shop, response, err := client.Shop.Get(r.Context())
					if err != nil {
						if response != nil && response.StatusCode >= 400 && response.StatusCode < 500 {
							place.Status = data.PlaceStatusInActive
							data.DB.Place.Save(place)
							a.Status = place.Status
						} else {
							lg.Warnf("place %d unhandled error %v", place.ID, err)
						}
					}
					if shop != nil {
						a.Email = shop.Email
						a.Country = shop.Country
						a.Phone = shop.Phone
					}

					a.ProductCount, _ = data.DB.Product.Find(db.Cond{"place_id": place.ID}).Count()

				}
				fmt.Printf(".")
				activeChann <- *a
			}
			sg.Done()
		}(credChann)
	}

	var creds []*data.ShopifyCred
	data.DB.Select(db.Raw("c.*")).
		From("shopify_creds c").
		LeftJoin("places p").On("p.id = c.place_id").
		Where(db.Cond{"p.status": data.PlaceStatusActive}).
		All(&creds)

	var places []activePlace
	go func(activePlace []activePlace) {
		for a := range activeChann {
			places = append(places, a)
		}
	}(places)

	for _, cred := range creds {
		credChann <- cred
	}
	close(credChann)

	sg.Wait()

	render.Respond(w, r, places)
}

func GetMerchantProductCount(w http.ResponseWriter, r *http.Request) {
	var creds []*data.ShopifyCred

	data.DB.Select(db.Raw("c.*")).
		From("shopify_creds c").
		LeftJoin("places p").On("p.id = c.place_id").
		Where(db.Cond{"p.status": data.PlaceStatusActive}).
		All(&creds)

	type countCompare struct {
		Theirs  int   `json:"theirs"`
		Ours    int   `json:"ours"`
		PlaceID int64 `json:"placeId"`
	}

	var result []countCompare
	for _, cred := range creds {
		client := shopify.NewClient(nil, cred.AccessToken)
		client.BaseURL, _ = url.Parse(cred.ApiURL)
		theirsCount, _, err := client.ProductList.Count(context.Background())
		if err != nil {
			log.Printf("place %d: %+v", cred.PlaceID, err)
			continue
		}
		ourCount, _ := data.DB.Product.Find(db.Cond{"place_id": cred.PlaceID}).Count()
		if err != nil {
			log.Printf("place %d: %+v", cred.PlaceID, err)
			continue
		}

		result = append(result, countCompare{
			Theirs:  int(theirsCount),
			Ours:    int(ourCount),
			PlaceID: cred.PlaceID,
		})
	}

	render.Respond(w, r, result)
}

func GetMerchantField() {
	var creds []*data.ShopifyCred

	data.DB.Select(db.Raw("c.*")).
		From("shopify_creds c").
		LeftJoin("places p").On("p.id = c.place_id").
		Where(db.Cond{"p.status": data.PlaceStatusActive}).
		All(&creds)

	for _, cred := range creds {
		client := shopify.NewClient(nil, cred.AccessToken)
		client.BaseURL, _ = url.Parse(cred.ApiURL)
		shop, _, err := client.Shop.Get(context.Background())
		if err != nil {
			continue
		}
		fmt.Printf("(%d,%s,%s),\n", cred.PlaceID, shop.PlanName, shop.PlanDisplayName)
	}
}

func GetShopifyPaymentMethods() {
	cond := db.And(
		db.Cond{"status": data.PlaceStatusActive},
		db.Or(
			db.Cond{"payment_methods": db.IsNull()},
			db.Cond{"payment_methods": "null"},
			db.Cond{"payment_methods": "[]"},
		),
	)

	var places []*data.Place
	data.DB.Place.Find(cond).All(&places)

	for _, p := range places {
		cred, err := data.DB.ShopifyCred.FindByPlaceID(p.ID)
		if err != nil {
			continue
		}
		client := shopify.NewClient(nil, cred.AccessToken)
		client.BaseURL, _ = url.Parse(cred.ApiURL)

		p.PaymentMethods = []*data.PaymentMethod{}
		if checkout, _, _ := client.Checkout.Create(context.Background(), nil); checkout != nil && len(checkout.ShopifyPaymentAccountID) != 0 {
			// NOTE: for now the id returned on checkout is stripe specific
			p.PaymentMethods = append(p.PaymentMethods, &data.PaymentMethod{Type: "stripe", ID: checkout.ShopifyPaymentAccountID})

			log.Printf("found payment method for place %d, %s", p.ID, p.Name)
			data.DB.Place.Save(p)
		}
		fmt.Printf(".")
	}
}

func GetCheckoutStatus() {
	var creds []*data.ShopifyCred
	data.DB.Select(db.Raw("c.*")).
		From("shopify_creds c").
		LeftJoin("places p").On("p.id = c.place_id").
		Where(db.Cond{"p.status": data.PlaceStatusActive}).
		All(&creds)

	clientMap := make(map[int64]*shopify.Client)
	for _, cred := range creds {
		client := shopify.NewClient(nil, cred.AccessToken)
		client.BaseURL, _ = url.Parse(cred.ApiURL)
		clientMap[cred.PlaceID] = client
	}

	carts, _ := data.DB.Cart.FindAll(db.Cond{"status": data.CartStatusCheckout})
	for _, c := range carts {
		for placeID, v := range c.Etc.ShopifyData {
			client, ok := clientMap[placeID]
			if !ok {
				log.Printf("place %d not found in client", placeID)
				continue
			}

			checkout, _, err := client.Checkout.Get(context.Background(), v.Token)
			if err != nil {
				log.Printf("place %d checkout %s with %v", placeID, v.Name, err)
				continue
			}

			if checkout.OrderID != 0 {
				log.Printf("%+v", checkout)
			}
		}
	}
}
