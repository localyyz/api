package tool

import (
	"context"
	"log"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/lib/sync"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
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

func syncUpdateProductWorker(ctx context.Context, jobs chan int, results chan int) {
	client := ctx.Value("shopify.client").(*shopify.Client)
	place := ctx.Value("place").(*data.Place)

	for pageNum := range jobs {
		var products []*data.Product
		data.DB.Product.Find(
			db.Cond{
				"place_id":    place.ID,
				"deleted_at":  db.Is(nil),
				"external_id": db.IsNot(nil),
			},
		).Limit(200).Offset(200 * (pageNum - 1)).All(&products)
		externalIDs := []int64{}
		for _, p := range products {
			// CHECK IF IMAGE IS GOOD
			//resp, err := http.DefaultClient.Head(p.ImageUrl)
			//if err != nil {
			//log.Printf("head error %d: %+v", p.ID, err)
			//continue
			//}
			//if resp.StatusCode != http.StatusNotFound {
			//continue
			//}
			externalIDs = append(externalIDs, *p.ExternalID)
		}
		if len(externalIDs) == 0 {
			log.Printf("nothing to do for page %d", pageNum)
			results <- 0
			continue
		}
		log.Printf("found %d products to update", len(externalIDs))

		productList, _, _ := client.ProductList.Get(
			ctx,
			&shopify.ProductListParam{
				Limit:      200,
				Page:       1,
				ProductIDs: externalIDs,
			},
		)
		if len(productList) <= 200 && len(productList) > len(externalIDs) {
			log.Printf("mismatch length got %d expected %d", len(productList), len(externalIDs))
			results <- 0
			continue
		}
		if len(productList) == 0 {
			log.Printf("no more pages at %d", pageNum)
			results <- 0
			continue
		}

		log.Printf("found %d to update", len(productList))
		ctx = context.WithValue(ctx, "sync.list", productList)
		sync.ShopifyProductListingsUpdate(ctx)

		results <- 1
		lg.Infof("Done page %d", pageNum)
	}
}

func syncCreateProductWorker(ctx context.Context, jobs chan int, results chan int) {
	client := ctx.Value("shopify.client").(*shopify.Client)
	for pageNum := range jobs {
		productList, _, _ := client.ProductList.Get(
			ctx,
			&shopify.ProductListParam{
				Limit: 200,
				Page:  pageNum,
			},
		)
		if len(productList) == 0 {
			log.Printf("no more pages at %d", pageNum)
			results <- 0
			continue
		}
		log.Printf("found %d to create", len(productList))
		ctx = context.WithValue(ctx, "sync.list", productList)
		if err := sync.ShopifyProductListingsCreate(ctx); err != nil {
			log.Println(err)
		}

		results <- 1
		lg.Infof("Done page %d", pageNum)
	}
}

func finalizeStatus(ctx context.Context, hasCategory bool, inputs ...string) data.ProductStatus {
	// if blacklisted, demote the product status
	if sync.SearchBlackList(ctx, inputs...) {
		if hasCategory {
			// mark as pending, blacklisted but found a category
			return data.ProductStatusPending
		} else {
			// reject if we did not find a category
			return data.ProductStatusRejected
		}
	}

	if hasCategory {
		return data.ProductStatusApproved
	}
	return data.ProductStatusPending
}

func BlacklistProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
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

	ctx = context.WithValue(ctx, "category.blacklist", blacklistCache)

	iterator := data.DB.Select("*").
		From("products").
		Where(db.Cond{
			"deleted_at": nil,
		}).
		Limit(200).
		Iterator()
	if err != nil {
		lg.Warn(err)
		return
	}
	for {
		var p *data.Product
		if !iterator.Next(&p) {
			lg.Debug("done")
			break
		}

		newStatus := finalizeStatus(ctx, len(p.Category.Value) > 0, p.Title)
		if newStatus != p.Status {
			lg.Infof("%d: %s -> %s", p.ID, newStatus, p.Title)
		}
	}

	iterator.Close()
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

	ctx = context.WithValue(ctx, "category.blacklist", blacklistCache)
	ctx = context.WithValue(ctx, "category.cache", categoryCache)
	ctx = context.WithValue(ctx, "sync.place", place)

	jobs := make(chan int, 10)
	results := make(chan int, 10)

	// Create or update depending on PUT or POST
	isCreate := r.Method == http.MethodPost
	for i := 0; i < 10; i++ {
		log.Printf("%s worker %d started.", r.Method, i)

		if isCreate {
			go syncCreateProductWorker(ctx, jobs, results)
		} else {
			go syncUpdateProductWorker(ctx, jobs, results)
		}

		ii := i + 1
		jobs <- ii
	}

	//max, _ := data.DB.Product.Find(db.Cond{"place_id": place.ID}).Count()
	doneCount := 0
	page := 10
W:
	for {
		select {
		case r := <-results:
			//case <-results:
			if r == 0 {
				doneCount += 1 // if all came back without results. end
				if doneCount == 10 {
					//if (page-1)*200 > int(max) {
					close(jobs)
					break W
				}
			}
			if r == 1 {
				page += 1
				log.Printf("starting new pages %d", page)
				jobs <- page
			}
		default:
			continue
		}
	}
	log.Printf("finished %d pages of work", page)
}

func UpdateCategories(w http.ResponseWriter, r *http.Request) {
	UpdateProductCategory(r.Context())
	render.Respond(w, r, "done")
}
