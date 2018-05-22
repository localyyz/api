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
	TOTALIMAGEWEIGHT  = 5
	MINIMUMIMAGEWIDTH = 750
)

type shopifyImageSyncer struct {
	Product   *data.Product
	toSaves   []*data.ProductImage
	toRemoves []*data.ProductImage
}

type shopifyImageScorer struct {
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

func (s *shopifyImageScorer) ScoreProduct(images []*data.ProductImage) (int64, error) {
	if len(images) == 0 {
		return -1, nil
	}

	for _, img := range images {
		res, _ := http.Head(img.ImageURL)
		if res.StatusCode != 200 { //if image is not valid error out
			return 0, errors.New("Error: 404 image url")
		}
	}

	var totalScore int64
	var pictureWeight float64

	pictureWeight = float64(TOTALIMAGEWEIGHT) / float64(len(images)) //the weight of each individual picture

	/* scoring each picture */
	for _, val := range images {
		if val.Width >= MINIMUMIMAGEWIDTH {
			totalScore++
			val.Score = 1
		} else {
			val.Score = 0
		}
	}

	productScore := int64(math.Ceil(pictureWeight * float64(totalScore))) //the product score from each image

	return productScore, nil
}

func setImages(syncer productImageSyncer, scorer productImageScorer, imgs ...*shopify.ProductImage) (int64, error) {

	//getting the images from product_images for product
	dbImages, err := syncer.FetchProductImages()
	if err != nil {
		return 0, err
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

	// score the images, return err if any image url is a 404
	productScore, err := scorer.ScoreProduct(toSaves)

	// if there is a 404 return the error from ScoreProduct, product is rejected anyways so no need to save
	if err != nil {
		return productScore, err
	} else {
		//no 404 images so save product and return score
		return productScore, syncer.Finalize(toSaves, toRemoves)
	}

}
