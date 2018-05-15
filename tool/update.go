package tool

import (
	"context"
	"net/url"
	"sync"
	"time"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
)

type productParse []*data.Product

func UpdateProductCategory(ctx context.Context) {
	/* creating category cache */
	categoryCache := make(map[string]*data.Category)
	if categories, _ := data.DB.Category.FindAll(nil); categories != nil {
		for _, c := range categories {
			categoryCache[c.Value] = c
		}
	}

	/* creating blacklist cache */
	blacklistCache := make(map[string]*data.Blacklist)
	if blacklist, _ := data.DB.Blacklist.FindAll(nil); blacklist != nil {
		for _, word := range blacklist {
			blacklistCache[word.Word] = word
		}
	}

	shopifyClientCache := make(map[int64]*shopify.Client)
	var creds []*data.ShopifyCred
	data.DB.Select(db.Raw("c.*")).
		From("shopify_creds c").
		LeftJoin("places p").On("p.id = c.place_id").
		Where(db.Cond{"p.status": data.PlaceStatusActive}).
		All(&creds)
	for _, cred := range creds {
		client := shopify.NewClient(nil, cred.AccessToken)
		client.BaseURL, _ = url.Parse(cred.ApiURL)
		shopifyClientCache[cred.PlaceID] = client
	}

	lg.Infof("category: %d", len(categoryCache))
	lg.Infof("blacklist: %d", len(blacklistCache))
	ctx = context.WithValue(ctx, "category.cache", categoryCache)
	ctx = context.WithValue(ctx, "category.blacklist", blacklistCache)
	ctx = context.WithValue(ctx, "sync.place", &data.Place{})
	ctx = context.WithValue(ctx, "shopify.cache", shopifyClientCache)

	productsChann := make(chan productParse, 10)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		lg.Printf("starting worker %d", i+1)
		wg.Add(1)
		go worker(ctx, productsChann, wg)
	}

	go func() {
		var (
			limit = 1000
			page  = 0
		)
		for {
			var products []*data.Product
			err := data.DB.Product.Find(
				db.Cond{
					"external_id": db.IsNot(nil),
					"status":      db.NotEq(data.ProductStatusRejected),
					"created_at":  db.Lt(time.Now().Add(-time.Hour)),
				},
			).
				Limit(limit).
				Offset(page * limit).
				OrderBy("id").
				All(&products)

			if err != nil {
				lg.Print("Error: could not load products")
				break
			}
			if len(products) == 0 {
				lg.Print("Finished")
				break
			}

			productsChann <- productParse(products)

			lg.Printf("sent page %d", page)
			page++
		}

		close(productsChann)
	}()

	// Wait for all HTTP fetches to complete.
	wg.Wait()
	lg.Info("done.")
}

func worker(ctx context.Context, ch chan productParse, wg sync.WaitGroup) {
	//var updated []*data.Product
	clientCache := ctx.Value("shopify.cache").(map[int64]*shopify.Client)
	for products := range ch {
		for _, product := range products {

			//if product.Category.Type == 0 {
			//parsedData := s.ParseProduct(ctx, product.Title, product.Description) //getting the category
			//if len(parsedData.Value) > 0 {
			//product.Category = data.ProductCategory{ //setting the category
			//Type:  parsedData.Type,
			//Value: parsedData.Value,
			//}
			//}
			//}

			//oldStatus := product.Status
			//product.Status = finalizeStatus(ctx, len(product.Category.Value) > 0, product.Title)
			//if product.Status == data.ProductStatusPending {
			//product.Status = data.ProductStatusRejected
			//}

			//if oldStatus == product.Status {
			//continue
			//}

			// find the product images
			client := clientCache[product.PlaceID]
			images, _, _ := client.Product.GetImages(ctx, *product.ExternalID)
			for _, img := range images {
				dbImage, err := data.DB.ProductImage.FindByExternalID(img.ID)
				if err != nil {
					continue
				}
				if dbImage.Width != img.Width || dbImage.Height != img.Height {
					dbImage.Width = img.Width
					dbImage.Height = img.Height
					data.DB.ProductImage.Save(dbImage)
				}
			}

			//saving it to the db
			//if err := data.DB.Product.Save(product); err != nil { //updating db
			//lg.Print("Error: Could not update the database entries")
			//}
			//updated = append(updated, product)
		}
		//for _, p := range updated {
		//lg.Printf("updated product (%d) to %s", p.ID, p.Status)
		//}
	}
	wg.Done()
}
