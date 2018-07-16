package tool

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	s "bitbucket.org/moodie-app/moodie-api/lib/sync"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	set "gopkg.in/fatih/set.v0"
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
	ourCount, _ := data.DB.Product.Find(db.Cond{
		"place_id":   place.ID,
		"deleted_at": db.IsNull(),
	}).Count()
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Respond(w, r, map[string]int{"theirs": int(theirsCount), "ours": int(ourCount)})
}

func CleanupProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	place := ctx.Value("place").(*data.Place)
	client := ctx.Value("shopify.client").(*shopify.Client)

	page := 1
	productSet := set.New()
	for {
		products, _, _ := client.ProductList.Get(
			ctx,
			&shopify.ProductListParam{
				Page:  page,
				Limit: 200,
			},
		)
		if len(products) == 0 {
			break
		}
		log.Printf("fetched page %d", page)
		for _, v := range products {
			productSet.Add(v.ProductID)
		}
		log.Printf("size %d", productSet.Size())
		page++
	}
	log.Println("done fetching products")

	removeSet := set.New()

	dbProducts, _ := data.DB.Product.FindAll(db.Cond{"place_id": place.ID})
	for _, v := range dbProducts {
		if !productSet.Has(v.ExternalID) {
			removeSet.Add(v.ID)
		}
	}

	log.Printf("removed %d products", removeSet.Size())

	data.DB.Product.Find(db.Cond{"id": set.IntSlice(removeSet)}).Delete()
}

func SyncProduct(w http.ResponseWriter, r *http.Request) {
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

	{
		externalID, err := strconv.ParseInt(chi.URLParam(r, "externalID"), 10, 64)
		if err != nil {
			log.Println("invalid param")
			return
		}

		// Create or update depending on PUT or POST
		p, _, _ := client.ProductList.GetProduct(
			ctx,
			externalID,
		)
		if p == nil {
			log.Println("not found")
			return
		}

		if !p.Available {
			return
		}
		ctx := context.WithValue(context.Background(), "sync.list", []*shopify.ProductList{p})
		ctx = context.WithValue(ctx, "category.blacklist", blacklistCache)
		ctx = context.WithValue(ctx, "category.cache", categoryCache)
		ctx = context.WithValue(ctx, "sync.place", place)
		s.ShopifyProductListingsUpdate(ctx)
	}

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

	for i := 0; i <= 34; i++ {
		log.Printf("fetching page %d", i+1)

		var products []*data.Product
		err := data.DB.Select("p.*").
			From("products p").
			LeftJoin("product_variants pv").On("pv.product_id = p.id").
			Where(db.Cond{
				"pv.place_id":             place.ID,
				"p.deleted_at":            db.IsNull(),
				"p.status":                data.ProductStatusApproved,
				db.Raw("pv.etc->>'size'"): "",
			}).
			Limit(200).
			Offset(i * 200).
			All(&products)

		if err != nil {
			log.Println(err)
			break
		}
		if len(products) == 0 {
			log.Println("db call got nothing")
			break
		}

		externalIDs := make([]int64, len(products))
		for i, v := range products {
			externalIDs[i] = *v.ExternalID
		}

		log.Println("Fetching %v", externalIDs)

		productList, _, _ := client.ProductList.Get(
			ctx,
			&shopify.ProductListParam{
				ProductIDs: externalIDs,
				Limit:      200,
				Page:       1,
			},
		)
		if len(productList) == 0 {
			log.Printf("no more pages at %d", i)
			break
		}
		log.Printf("found %d to create for page %d", len(productList), i+1)

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

func ValidateSyncProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)

	iter := data.DB.Select("*").
		From("products").
		Where(
			db.Cond{
				"place_id": place.ID,
				"status":   3,
				"id":       db.Gt(261303),
			},
		).
		OrderBy("id").
		Iterator()
	defer iter.Close()

	worker := func(i int, chnn chan data.Product) {
		for {
			p, ok := <-chnn
			if !ok {
				log.Printf("wrk(%d) finished.", i)
				break
			}

			pi, err := data.DB.ProductImage.FindOne(db.Cond{"product_id": p.ID})
			if err != nil {
				log.Println("wkr(%d): product_image %v", i, err)
				continue
			}

			s := time.Now()
			resp, err := http.DefaultClient.Head(pi.ImageURL)
			if err != nil {
				log.Printf("wrk(%d): product %d fetch image err %v", i, p.ID, err)
				continue
			}
			log.Printf("timeit: %v", time.Since(s))

			if resp.StatusCode == http.StatusNotFound {
				p.Status = data.ProductStatusRejected
				data.DB.Product.Save(&p)
				log.Printf("wrk(%d): product %d was rejected, timeit: %v", i, p.ID, time.Since(s))
			}
		}
	}

	productChann := make(chan data.Product, 10)
	for i := 1; i <= 10; i++ {
		log.Printf("starting worker %d", i)
		go worker(i, productChann)
	}

	for {
		var p data.Product
		if !iter.Next(&p) {
			log.Println("done")
			break
		}
		productChann <- p
	}

	if err := iter.Err(); err != nil {
		log.Println(err)
	}

	close(productChann)

}

func UpdateCategories(w http.ResponseWriter, r *http.Request) {
	UpdateProductCategory(r.Context())
	render.Respond(w, r, "done")
}
