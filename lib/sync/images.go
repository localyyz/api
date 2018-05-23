package sync

import (
	"errors"
	"math"
	"net/http"
	"net/url"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"github.com/pressly/lg"
	set "gopkg.in/fatih/set.v0"
	db "upper.io/db.v3"
)

const (
	TotalImageWeight  = 5
	MinimumImageWidth = 750
)

type shopifyImageSyncer struct {
	Product   *data.Product
	toSaves   []*data.ProductImage
	toRemoves []*data.ProductImage
}

type shopifyImageScorer struct {
	Product *data.Product
}

// fetches existing product images from the database.
// function is abstracted so the db call can be mocked/tested
func (s *shopifyImageSyncer) FetchProductImages() ([]*data.ProductImage, error) {
	return data.DB.ProductImage.FindByProductID(s.Product.ID)
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

func (s *shopifyImageScorer) GetProduct() *data.Product {
	return s.Product
}

func (s *shopifyImageScorer) ScoreProductImages(images []*data.ProductImage) error {

	for _, img := range images {
		res, _ := http.Head(img.ImageURL)
		if res.StatusCode != 200 { //if image is not valid error out
			return errors.New("Error: 404 image url")
		}

		if img.Width >= MinimumImageWidth {
			img.Score = 1
		} else {
			img.Score = 0
		}
	}

	return nil
}

func (s *shopifyImageScorer) Finalize(images []*data.ProductImage) error {

	if len(images) == 0 {
		s.GetProduct().Score = 0
		return nil
	}

	var totalScore int64

	//the weight of each individual picture
	pictureWeight := float64(TotalImageWeight) / float64(len(images))
	for _, img := range images {
		totalScore += img.Score
	}

	//the product score from each image
	s.GetProduct().Score = int64(math.Ceil(pictureWeight * float64(totalScore)))
	return nil

}

func setImages(syncer productImageSyncer, scorer productImageScorer, imgs ...*shopify.ProductImage) error {

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
	var toKeeps []*data.ProductImage

	for _, img := range imgs {
		if syncImagesSet.Has(img.ID) {
			// duplicate image, pass
			continue
		}

		// image external id set needs to be synced
		syncImagesSet.Add(img.ID)

		if ext, ok := dbImagesMap[img.ID]; ok {
			// ignored. image already saved
			toKeeps = append(toKeeps, ext)
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

	// score the images, return err if any image url is a 404
	keepAndSave := append(toKeeps, toSaves...)
	err = scorer.ScoreProductImages(keepAndSave)
	if err != nil {
		return err
	}

	//aggregate the score
	scorer.Finalize(keepAndSave)

	//save the images
	err = syncer.Finalize(toSaves, toRemoves)
	if err != nil {
		return err
	}

	return nil

}
