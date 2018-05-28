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

type shopifyImageSyncer struct {
	Client    HttpClient
	Product   *data.Product
	toSaves   []*data.ProductImage
	toRemoves []*data.ProductImage
}

// fetches existing product images from the database.
// function is abstracted so the db call can be mocked/tested
func (s *shopifyImageSyncer) FetchProductImages() ([]*data.ProductImage, error) {
	return data.DB.ProductImage.FindByProductID(s.Product.ID)
}

func (s *shopifyImageSyncer) ValidateImages() bool {
	for _, val := range s.toSaves {
		req, err := http.NewRequest("HEAD", val.ImageURL, nil)
		if err != nil {
			lg.Warnf("Error: Could not create http request for image id: %d", val.ID)
		}
		res, err := s.Client.Do(req)
		if err != nil {
			lg.Warnf("Error: Could not load image url for image id: %d", val.ID)
		}
		if res.StatusCode != 200 {
			return false
		}
	}
	return true
}

func (s *shopifyImageSyncer) GetProduct() *data.Product {
	return s.Product
}

func (s *shopifyImageSyncer) Finalize(toSaves, toRemoves []*data.ProductImage) error {
	for _, img := range toSaves {

		data.DB.ProductImage.Save(img)

		// if img has variant ids associated, save to pivot table
		for _, vID := range img.VariantIDs {
			var variant *data.ProductVariant
			err := data.DB.ProductVariant.Find(db.Cond{
				"offer_id": vID,
			}).Select("id").One(&variant)
			if err != nil {
				lg.Warnf("failed to fetch variant(%d) with %+v", vID, err)
				continue
			}
			err = data.DB.VariantImage.Create(data.VariantImage{
				VariantID: variant.ID,
				ImageID:   img.ID,
			})
			if err != nil {
				lg.Warnf("failed to save variant image with %+v", err)
			}
		}
	}

	for _, img := range toRemoves {
		data.DB.ProductImage.Delete(img)
	}

	return nil
}

func setImages(syncer productImageSyncer, imgs ...*shopify.ProductImage) error {

	//getting the images from product_images for product
	dbImages, err := syncer.FetchProductImages()
	if err != nil {
		return err
	}

	//fill out a map using external image IDS
	dbImagesMap := map[int64]*data.ProductImage{}
	for _, img := range dbImages {
		dbImagesMap[img.ExternalID] = img
	}

	syncImagesSet := set.New()
	var toSaves []*data.ProductImage

	for _, img := range imgs {
		if syncImagesSet.Has(img.ID) {
			// duplicate image, pass
			continue
		}

		// image external id set needs to be synced
		syncImagesSet.Add(img.ID)

		if _, ok := dbImagesMap[img.ID]; ok {
			// ignored. image already saved
			continue
		}

		imgUrl, _ := url.Parse(img.Src)
		imgUrl.Scheme = "https"
		// remove any query params
		imgUrl.RawQuery = ""
		toSaves = append(toSaves, &data.ProductImage{
			ProductID:  syncer.GetProduct().ID,
			ExternalID: img.ID,
			ImageURL:   imgUrl.String(),
			Ordering:   int32(img.Position),
			VariantIDs: img.VariantIds,
			Width:      img.Width,
			Height:     img.Height,
		})
	}

	// for images not in update, remove them
	var toRemoves []*data.ProductImage
	for _, toRemove := range dbImagesMap {
		if !syncImagesSet.Has(toRemove.ExternalID) {
			toRemoves = append(toRemoves, toRemove)
		}
	}

	if !syncer.ValidateImages() {
		return errors.New("Error: 404 image url")
	}

	//save the images
	return syncer.Finalize(toSaves, toRemoves)
}
