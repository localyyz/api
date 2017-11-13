package main

import (
	"context"
	"log"
	"net/url"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
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
	pullProductCollection()
}

func pullProductCollection() {
	var creds []*data.ShopifyCred
	if err := data.DB.ShopifyCred.Find(db.Cond{"place_id": 15}).All(&creds); err != nil {
		log.Fatal(err)
	}

	for _, cred := range creds {
		cl := shopify.NewClient(nil, cred.AccessToken)
		cl.BaseURL, _ = url.Parse(cred.ApiURL)

		// find the products
		products, err := data.DB.Product.FindAll(db.Cond{"place_id": cred.PlaceID, "external_id IS NOT": nil})
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("found %d products for place %d\n", len(products), cred.PlaceID)

		// for each product, find the collections they belong to
		for _, p := range products {
			ctx := context.Background()
			clist, _, _ := cl.CustomCollection.Get(ctx, &shopify.CustomCollectionParam{ProductID: *p.ExternalID})
			log.Printf("found %d collections for product %d\n", len(clist), p.ID)

			// for each clist, parse the handle as tags
			for _, c := range clist {
				log.Println(c.Handle)
			}
		}

		break
	}
}

func pullExternalID() {
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
				p.ExternalID = &plist[0].ProductID
				data.DB.Product.Save(p)
			}
		}
	}
}
