package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/lib/sync"
	"github.com/gedex/inflector"
	"github.com/pressly/lg"
	set "gopkg.in/fatih/set.v0"
	db "upper.io/db.v3"
)

func main() {
	conf := data.DBConf{
		Database:        "localyyz",
		Hosts:           []string{"localhost"},
		Username:        "localyyz",
		ApplicationName: "sync_tool",
	}
	if _, err := data.NewDBSession(&conf); err != nil {
		log.Fatal(err)
	}
	log.Println("tool started.")

	pullProducts()
	//pullProductExternalID()
	//pullProductGender()
	//pullProductCategory()
}

var tagRegex = regexp.MustCompile("[^a-zA-Z0-9-]+")

func parseTags(tagStr string, optTags ...string) []string {
	tt := tagRegex.Split(tagStr, -1)

	tagSet := set.New()
	for _, t := range tt {
		t = strings.ToLower(t)

		tt := inflector.Singularize(t)
		for {
			if tt == t {
				break
			}
			t = tt
			tt = inflector.Singularize(t)
		}
		if tt == "" {
			continue
		}
		tagSet.Add(t)
	}
	for _, t := range optTags {
		tagSet.Add(strings.ToLower(t))
	}

	return set.StringSlice(tagSet)
}

func noGender(v string) bool {
	return !(v == "man" || v == "woman")
}

func filter(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func pullProductCategory() {
	var creds []*data.ShopifyCred
	err := data.DB.ShopifyCred.Find(
		db.Cond{
			"place_id": 44,
		},
	).All(&creds)
	if err != nil {
		log.Fatal(err)
	}

	for _, cred := range creds {
		cl := shopify.NewClient(nil, cred.AccessToken)
		cl.BaseURL, _ = url.Parse(cred.ApiURL)

		// find the products
		products, err := data.DB.Product.FindAll(
			db.Cond{
				"place_id":           cred.PlaceID,
				"external_id IS NOT": nil,
			},
		)
		if err != nil {
			log.Fatal(err)
		}

		// for each product, find the collections they belong to
		//q := data.DB.InsertInto("product_tags").
		//Columns("product_id", "place_id", "value", "type").
		//Amend(func(query string) string {
		//return query + ` ON CONFLICT DO NOTHING`
		//})
		//b := q.Batch(len(products))

		//go func() {
		//defer b.Done()
		for _, p := range products {
			ctx := context.Background()
			clist, _, _ := cl.CustomCollection.Get(ctx, &shopify.CustomCollectionParam{ProductID: *p.ExternalID})

			// for each clist, parse the title as tags
			for _, c := range clist {
				tt := filter(parseTags(c.Title), noGender)
				if len(tt) > 0 {
					fmt.Println(c.Title, tt)
				}
			}
			//}
			//}()

			//if err := b.Wait(); err != nil {
			//lg.Warn(err)
			//}

		}
	}

}

func pullProductGender() {
	var creds []*data.ShopifyCred
	err := data.DB.ShopifyCred.Find(
		db.Cond{
			"place_id": 20,
		},
	).All(&creds)
	if err != nil {
		log.Fatal(err)
	}

	for _, cred := range creds {
		cl := shopify.NewClient(nil, cred.AccessToken)
		cl.BaseURL, _ = url.Parse(cred.ApiURL)

		// find the products
		products, err := data.DB.Product.FindAll(
			db.Cond{
				"place_id":           cred.PlaceID,
				"external_id IS NOT": nil,
			},
		)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("found %d products for place %d\n", len(products), cred.PlaceID)

		// for each product, find the collections they belong to

		q := data.DB.InsertInto("product_tags").
			Columns("product_id", "place_id", "value", "type").
			Amend(func(query string) string {
				return query + ` ON CONFLICT DO NOTHING`
			})
		b := q.Batch(len(products))

		go func() {
			defer b.Done()
			for _, p := range products {
				ctx := context.Background()
				clist, _, _ := cl.CustomCollection.Get(ctx, &shopify.CustomCollectionParam{ProductID: *p.ExternalID})

				// for each clist, parse the handle as tags
				for _, c := range clist {
					t := strings.ToLower(c.Handle)
					tt := inflector.Singularize(t)
					for {
						if tt == t {
							break
						}
						t = tt
						tt = inflector.Singularize(t)
					}
					if tt == "" {
						continue
					}
					if tt == "man" || tt == "woman" {
						b.Values(p.ID, p.PlaceID, tt, data.ProductTagTypeGender)
					}
				}
			}
		}()

		if err := b.Wait(); err != nil {
			lg.Warn(err)
		}

		break
	}
}

func pullProductExternalID() {
	var creds []*data.ShopifyCred
	if err := data.DB.ShopifyCred.Find().All(&creds); err != nil {
		log.Fatal(err)
	}

	for _, cred := range creds {
		cl := shopify.NewClient(nil, cred.AccessToken)
		cl.BaseURL, _ = url.Parse(cred.ApiURL)

		// find the products
		products, err := data.DB.Product.FindAll(db.Cond{"place_id": cred.PlaceID, "external_id": nil})
		if err != nil {
			log.Fatal(err)
		}

		for _, p := range products {
			// range of products and fill external_id
			ctx := context.Background()
			plist, _, _ := cl.ProductList.Get(ctx, &shopify.ProductListParam{Handle: p.ExternalHandle})
			if len(plist) > 0 {
				_, err := data.DB.Update("products").
					Set("external_id", plist[0].ProductID).
					Where(db.Cond{"id": p.ID}).
					Exec()
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}

func pullProducts() {
	var creds []*data.ShopifyCred
	if err := data.DB.ShopifyCred.Find().All(&creds); err != nil {
		log.Fatal(err)
	}

	for _, cred := range creds {
		cl := shopify.NewClient(nil, cred.AccessToken)
		cl.BaseURL, _ = url.Parse(cred.ApiURL)

		place, err := data.DB.Place.FindByID(cred.PlaceID)
		if err != nil {
			log.Fatal(err)
		}

		// range of products and fill external_id
		ctx := context.Background()
		plist, _, err := cl.ProductList.Get(ctx, nil)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("syncing %d products for %s", len(plist), place.Name)
		ctx = context.WithValue(ctx, "sync.list", plist)
		ctx = context.WithValue(ctx, "sync.place", place)

		sync.ShopifyProductListings(ctx)
	}

}
