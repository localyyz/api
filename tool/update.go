package tool

import (
	"context"
	"net/url"
	"sync"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
)

type imageParse []*shopify.ProductImage

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

	var clients []*shopify.Client
	var creds []*data.ShopifyCred
	data.DB.Select(db.Raw("c.*")).
		From("shopify_creds c").
		LeftJoin("places p").On("p.id = c.place_id").
		Where(db.Cond{"p.status": data.PlaceStatusActive}).
		All(&creds)
	for _, cred := range creds {
		client := shopify.NewClient(nil, cred.AccessToken)
		client.BaseURL, _ = url.Parse(cred.ApiURL)
		clients = append(clients, client)
	}

	//lg.Infof("category: %d", len(categoryCache))
	//lg.Infof("blacklist: %d", len(blacklistCache))
	//ctx = context.WithValue(ctx, "category.cache", categoryCache)
	//ctx = context.WithValue(ctx, "category.blacklist", blacklistCache)
	//ctx = context.WithValue(ctx, "sync.place", &data.Place{})
	//ctx = context.WithValue(ctx, "shopify.cache", shopifyClientCache)

	imagesChann := make(chan imageParse, 10)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		lg.Printf("starting worker %d", i+1)
		wg.Add(1)
		go worker(ctx, imagesChann, wg)
	}

	go func() {
		for i, cl := range clients {
			var (
				page   = 0
				limits = 250
			)
			for {
				products, _, _ := cl.ProductList.Get(ctx, &shopify.ProductListParam{Limit: limits, Page: page})
				if len(products) == 0 {
					break
				}

				for _, p := range products {
					imagesChann <- imageParse(p.Images)
				}

				page++
			}
			lg.Printf("sent page client %d", i)
		}
		close(imagesChann)
	}()

	// Wait for all HTTP fetches to complete.
	wg.Wait()
	lg.Info("done.")
}

func worker(ctx context.Context, ch chan imageParse, wg sync.WaitGroup) {
	//var updated []*data.Product
	for images := range ch {
		for _, image := range images {

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
			dbImage, err := data.DB.ProductImage.FindByExternalID(image.ID)
			if err != nil {
				continue
			}
			if dbImage.Width != image.Width || dbImage.Height != image.Height {
				dbImage.Width = image.Width
				dbImage.Height = image.Height
				data.DB.ProductImage.Save(dbImage)
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
