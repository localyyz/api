package sync

import (
	"net/http"
	"net/url"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"github.com/pkg/errors"
	"github.com/pressly/lg"
	set "gopkg.in/fatih/set.v0"
	db "upper.io/db.v3"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type shopifyImageSyncer struct {
	Product   *data.Product
	toSaves   []*data.ProductImage
	toRemoves []*data.ProductImage
	dbImages  []*data.ProductImage

	// interface for fetching http
	HTTPClient
	// interface for finalizing
	Finalizer
	// interface for fetching
	Fetcher
}

type shopifyImageFinalizer struct {
	toSaves   []*data.ProductImage
	toRemoves []*data.ProductImage
}

type shopifyImageFetcher struct{}

func (s *shopifyImageFinalizer) Finalize() error {
	for _, img := range s.toSaves {
		data.DB.ProductImage.Save(img)

		// if img has variant ids associated, save to pivot table
		for _, vID := range img.VariantIDs {
			var variant *data.ProductVariant
			err := data.DB.ProductVariant.Find(
				db.Cond{"offer_id": vID},
			).Select("id").One(&variant)
			if err != nil {
				return err
			}
			err = data.DB.VariantImage.Create(data.VariantImage{
				VariantID: variant.ID,
				ImageID:   img.ID,
			})
			if err != nil {
				return err
			}
		}
	}
	for _, img := range s.toRemoves {
		if err := data.DB.ProductImage.Delete(img); err != nil {
			return err
		}
	}
	return nil
}

func (s *shopifyImageSyncer) Finalize() error {
	if s.Finalizer == nil {
		s.Finalizer = &shopifyImageFinalizer{
			toSaves:   s.toSaves,
			toRemoves: s.toRemoves,
		}
	}
	return s.Finalizer.Finalize()
}

var (
	ErrInvalidImage = errors.New("invalid image")
	ErrEmptyImage   = errors.New("empty image")
)

// fetches existing product images from the database.
// function is abstracted so the db call can be mocked/tested
func (s *shopifyImageFetcher) Fetch(cond db.Cond, sliceOfStructs interface{}) error {
	err := data.DB.ProductImage.Find(cond).All(sliceOfStructs)
	return err
}

func (s *shopifyImageSyncer) ValidateImages() bool {
	for _, val := range s.toSaves {
		req, err := http.NewRequest("HEAD", val.ImageURL, nil)
		if err != nil {
			lg.Warnf("Error: Could not create http request for image id: %d", val.ID)
		}

		if s.HTTPClient == nil {
			s.HTTPClient = http.DefaultClient
		}

		res, err := s.Do(req)
		if err != nil {
			lg.Warnf("Error: Could not load image url for image id: %d", val.ID)
		}
		if res.StatusCode != http.StatusOK {
			return false
		}
	}
	return true
}

func (s *shopifyImageSyncer) GetProduct() *data.Product {
	return s.Product
}

func getScore(img *shopify.ProductImage) int64 {
	if img.Width >= MinimumImageWidth {
		return 1
	}
	return 0
}

func (s *shopifyImageSyncer) Sync(imgs []*shopify.ProductImage) error {
	if len(imgs) == 0 {
		return ErrEmptyImage
	}

	if s.Fetcher == nil {
		s.Fetcher = &shopifyImageFetcher{}
	}
	//getting the images from product_images for product
	var dbImages []*data.ProductImage
	if err := s.Fetch(db.Cond{"product_id": s.Product.ID}, &dbImages); err != nil {
		return errors.Wrap(err, "fetch")
	}

	//fill out a map using external image IDS
	dbImagesMap := map[int64]*data.ProductImage{}
	for _, img := range dbImages {
		dbImagesMap[img.ExternalID] = img
	}

	// set of images that should be kept
	syncImagesSet := set.New()
	for _, img := range imgs {
		// add to image sets.
		syncImagesSet.Add(img.ID)

		// ignored. image already in database
		if _, ok := dbImagesMap[img.ID]; ok {
			continue
		}

		imgUrl, _ := url.Parse(img.Src)
		imgUrl.Scheme = "https"

		s.toSaves = append(s.toSaves, &data.ProductImage{
			ProductID:  s.Product.ID,
			ExternalID: img.ID,
			ImageURL:   imgUrl.String(),
			Ordering:   int32(img.Position),
			VariantIDs: img.VariantIds,
			Width:      img.Width,
			Height:     img.Height,
			Score:      getScore(img),
		})
	}

	// for the images in database but not in the
	// shopify images list. remove them
	for _, img := range dbImagesMap {
		if !syncImagesSet.Has(img.ExternalID) {
			s.toRemoves = append(s.toRemoves, img)
		}
	}

	if !s.ValidateImages() {
		return ErrInvalidImage
	}

	//save or remove the images
	return s.Finalize()
}
