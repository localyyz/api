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

type imageList []*data.ProductImage

type workPackage struct {
	client   *shopify.Client
	page     int
	imagechn chan imageList
}

func UpdateProductCategory(ctx context.Context) {
	workpackageChan := make(chan workPackage)
	imagechn := make(chan imageList)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go worker(ctx, workpackageChan, wg)
		wg.Add(1)
		go saver(ctx, imagechn, wg)
	}

	go func() {
		var creds []*data.ShopifyCred
		data.DB.Select(db.Raw("c.*")).
			From("shopify_creds c").
			LeftJoin("places p").On("p.id = c.place_id").
			Where(db.Cond{"p.status": data.PlaceStatusActive}).
			All(&creds)

		for _, cred := range creds {
			client := shopify.NewClient(nil, cred.AccessToken)
			client.BaseURL, _ = url.Parse(cred.ApiURL)

			count, _, _ := client.ProductList.Count(ctx)
			lg.Infof("place (%d) found %d products", cred.PlaceID, count)
			var page int

			for i := 0; i < count; i += 20 {
				workpackageChan <- workPackage{
					client:   client,
					page:     page,
					imagechn: imagechn,
				}
				page += 1
			}
			lg.Infof("done place (%d)", cred.PlaceID)
		}

		close(workpackageChan)
		close(imagechn)
	}()

	// Wait for all HTTP fetches to complete.
	wg.Wait()
	lg.Info("done.")
}

func saver(ctx context.Context, list chan imageList, wg sync.WaitGroup) {
	for l := range list {
		lg.Infof("updating %d images", len(l))
		for _, img := range l {
			data.DB.ProductImage.Save(img)
		}
	}
	wg.Done()
}

func worker(ctx context.Context, packages chan workPackage, wg sync.WaitGroup) {
	for w := range packages {
		s := time.Now()
		products, _, _ := w.client.ProductList.Get(
			ctx,
			&shopify.ProductListParam{
				Limit: 20,
				Page:  w.page,
			},
		)
		//lg.Infof("fetched in %v", time.Since(s))
		if len(products) == 0 {
			continue
		}

		externalIDs := []int64{}
		imageMap := map[int64]*shopify.ProductImage{}
		for _, p := range products {
			for _, image := range p.Images {
				externalIDs = append(externalIDs, image.ID)
				imageMap[image.ID] = image
			}
		}

		dbImages, err := data.DB.ProductImage.FindAll(db.Cond{"external_id": externalIDs})
		if err != nil {
			continue
		}

		var updates imageList
		for _, img := range dbImages {
			ext := imageMap[img.ExternalID]
			if img.Width != ext.Width || img.Height != ext.Height {
				img.Width = ext.Width
				img.Height = ext.Height
				updates = append(updates, img)
			}
		}

		if len(updates) > 0 {
			lg.Infof("sending %d updates in %v", len(updates), time.Since(s))
			w.imagechn <- updates
		}
	}
	wg.Done()
}
