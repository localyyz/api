package tool

import (
	"log"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	set "gopkg.in/fatih/set.v0"
	db "upper.io/db.v3"
)

func SyncVariants(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	place := ctx.Value("place").(*data.Place)
	client := ctx.Value("shopify.client").(*shopify.Client)

	page := 1
	variantSet := set.New()
	// fetch all the variants and add them to variant set
	for {
		variants, _, _ := client.Variant.Get(ctx, &shopify.VariantParam{Page: page, Limit: 200})
		if len(variants) == 0 {
			break
		}
		log.Printf("fetched page %d", page)
		for _, v := range variants {
			variantSet.Add(v.ID)
		}
		log.Println("size %d", variantSet.Size())
		page++
	}
	log.Println("done fetching variants")

	dbVariants, _ := data.DB.ProductVariant.FindAll(db.Cond{"place_id": place.ID})
	var c int
	for _, v := range dbVariants {
		if !variantSet.Has(v.OfferID) {
			data.DB.ProductVariant.Delete(v)
			c += 1
		}
	}
	log.Printf("removed %d variants", c)
}
