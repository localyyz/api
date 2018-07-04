package tool

import (
	"context"
	"log"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	s "bitbucket.org/moodie-app/moodie-api/lib/sync"
	"github.com/go-chi/render"
	db "upper.io/db.v3"
)

func GetProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	client := ctx.Value("shopify.client").(*shopify.Client)

	query := r.URL.Query()
	params := &shopify.ProductListParam{
		Handle: query.Get("handle"),
	}

	product, _, err := client.ProductList.Get(ctx, params)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Respond(w, r, product)
}

func GetProductCount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	client := ctx.Value("shopify.client").(*shopify.Client)
	place := ctx.Value("place").(*data.Place)

	theirsCount, _, err := client.ProductList.Count(context.Background())
	if err != nil {
		render.Respond(w, r, err)
		return
	}
	ourCount, _ := data.DB.Product.Find(db.Cond{"place_id": place.ID}).Count()
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Respond(w, r, map[string]int{"theirs": int(theirsCount), "ours": int(ourCount)})
}

func SyncProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)

	categoryCache := make(map[string]*data.Category)
	categories, err := data.DB.Category.FindAll(nil)
	if err != nil {
		log.Print(err)
		return
	}
	for _, c := range categories {
		categoryCache[c.Value] = c
	}

	blacklistCache := make(map[string]*data.Blacklist)
	if blacklist, _ := data.DB.Blacklist.FindAll(nil); blacklist != nil {
		for _, word := range blacklist {
			blacklistCache[word.Word] = word
		}
	}
	client := ctx.Value("shopify.client").(*shopify.Client)

	// Create or update depending on PUT or POST

	for i := 1; i <= 30; i++ {
		log.Printf("fetching page %d", i)

		productList, _, _ := client.ProductList.Get(
			ctx,
			&shopify.ProductListParam{
				Limit: 200,
				Page:  i,
			},
		)
		if len(productList) == 0 {
			log.Printf("no more pages at %d", i)
			return
		}
		log.Printf("found %d to create for page %d", len(productList), i)

		for _, p := range productList {
			if !p.Available {
				continue
			}
			ctx := context.WithValue(context.Background(), "sync.list", []*shopify.ProductList{p})
			ctx = context.WithValue(ctx, "category.blacklist", blacklistCache)
			ctx = context.WithValue(ctx, "category.cache", categoryCache)
			ctx = context.WithValue(ctx, "sync.place", place)
			s.ShopifyProductListingsUpdate(ctx)
		}
	}

}

func UpdateCategories(w http.ResponseWriter, r *http.Request) {
	UpdateProductCategory(r.Context())
	render.Respond(w, r, "done")
}
