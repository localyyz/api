package sync

import (
	"math"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/pkg/errors"
	db "upper.io/db.v3"
)

const (
	TotalImageWeight      = 5
	MinimumImageWidth     = 750
	PriorityMerchantScore = 1
	CurrentyNotUSDScore   = -1
)

type shopifyImageScorer struct {
	Product *data.Product
	Place   *data.Place
}

func (s *shopifyImageScorer) GetProductImages(ID int64) ([]*data.ProductImage, error) {
	var images []*data.ProductImage
	if err := data.DB.ProductImage.Find(db.Cond{"product_id": ID}).All(&images); err != nil {
		return nil, errors.New("Error: Could not load product images from db")
	}
	return images, nil
}

func (s *shopifyImageScorer) GetProduct() *data.Product {
	return s.Product
}

func (s *shopifyImageScorer) GetPlace() *data.Place {
	return s.Place
}

func (s *shopifyImageScorer) Finalize(imgs []*data.ProductImage) error {
	if err := data.DB.ProductImage.Save(imgs); err != nil {
		return err
	} else {
		return nil
	}
}

func (s *shopifyImageScorer) CheckPriority() bool {
	return s.GetPlace().IsPriority()
}

func scoreIndividualImages(imgs []*data.ProductImage) {
	for _, img := range imgs {
		if img.Width >= MinimumImageWidth {
			img.Score = 1
		} else {
			img.Score = 0
		}
	}
}

func aggregateImageScore(imgs []*data.ProductImage) int64 {

	if len(imgs) == 0 {
		return 0
	}

	var totalScore int64

	//the weight of each individual picture
	pictureWeight := float64(TotalImageWeight) / float64(len(imgs))
	for _, img := range imgs {
		totalScore += img.Score
	}

	//the product score from each image
	return int64(math.Ceil(pictureWeight * float64(totalScore)))

}

func finalize(imgs []*data.ProductImage) error {
	return data.DB.ProductImage.Save(imgs)
}

func scoreProduct(scorer productScorer) error {

	if scorer.GetProduct().Status == data.ProductStatusRejected {
		scorer.GetProduct().Score = 0
		return nil
	}

	var imgScore int64
	var priorityMerchantScore int64
	var currencyNotUsdScore int64

	imgs, err := scorer.GetProductImages(scorer.GetProduct().ID)
	if err != nil {
		return err
	}
	scoreIndividualImages(imgs)
	imgScore = aggregateImageScore(imgs)

	if scorer.CheckPriority() {
		priorityMerchantScore = PriorityMerchantScore
	}

	if scorer.GetPlace().IsNotUSD() {
		currencyNotUsdScore = CurrentyNotUSDScore
	}

	if err = scorer.Finalize(imgs); err != nil {
		return err
	}

	scorer.GetProduct().Score = imgScore + priorityMerchantScore + currencyNotUsdScore
	return nil

}
