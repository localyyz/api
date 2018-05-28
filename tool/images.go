package tool

import (
	"context"
	"errors"
	"net/http"
	s "sync"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/sync"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
)

var (
	priorityMerchantIDs = []int64{
		121,
		116,
		606,
		2040,
		51,
		90,
		122,
		50,
		2713,
		887,
		1210,
		55,
		1671,
		58,
		678,
		773,
		3092,
	}
)

type toolScorer struct {
	Product *data.Product
	Place   *data.Place
}

func (s *toolScorer) GetProductImages(ID int64) ([]*data.ProductImage, error) {
	var images []*data.ProductImage
	if err := data.DB.ProductImage.Find(db.Cond{"product_id": ID}).All(&images); err != nil {
		return nil, errors.New("Error: Could not load product images from db")
	}
	return images, nil
}

func (s *toolScorer) GetProduct() *data.Product {
	return s.Product
}

func (s *toolScorer) GetPlace() *data.Place {
	return s.Place
}

func (s *toolScorer) Finalize(imgs []*data.ProductImage) error {
	for _, im := range imgs {
		if err := data.DB.ProductImage.Save(im); err != nil {
			return err
		}
	}
	return nil
}

func (s *toolScorer) CheckPriority() bool {
	return s.Place.IsPriority()
}

func scoreWorker(ctx context.Context, chn chan scoreProductList, wg s.WaitGroup) {
	defer wg.Done()

	placeCache := ctx.Value("place.cache").(map[int64]*data.Place)

	for pl := range chn {
		for _, p := range pl {
			sync.ScoreProduct(&toolScorer{Product: p, Place: placeCache[p.PlaceID]})
			data.DB.Product.Save(p)
		}
	}
}

type scoreProductList []*data.Product

func syncProductImageScores(w http.ResponseWriter, r *http.Request) {

	ctx := context.Background()
	productschn := make(chan scoreProductList)

	placeCache := map[int64]*data.Place{}
	places, _ := data.DB.Place.FindAll(db.Cond{"id": priorityMerchantIDs})
	for _, p := range places {
		placeCache[p.ID] = p
	}
	ctx = context.WithValue(ctx, "place.cache", placeCache)

	var wg s.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go scoreWorker(ctx, productschn, wg)
	}

	go func() {

		var offset int
		for {
			var products scoreProductList
			data.DB.Product.
				Find(
					db.Cond{
						"status":     data.ProductStatusApproved,
						"score":      db.Eq(-1),
						"deleted_at": db.Is(nil),
						"place_id":   priorityMerchantIDs,
					},
				).Limit(10).
				Offset(offset).
				All(&products)

			if len(products) == 0 {
				lg.Warn("we're done")
				break
			}

			productschn <- products
			offset += 10
			lg.Warnf("done %d", offset)

		}

		close(productschn)
	}()

	wg.Wait()
}
