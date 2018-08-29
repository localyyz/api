package sync

import (
	"math"

	"bitbucket.org/moodie-app/moodie-api/data"
)

const (
	TotalImageWeight      = 5
	MinimumImageWidth     = 750
	ScorePriorityMerchant = 1
	ScoreCurrencyPenalty  = -1
)

type shopifyProductScorer struct {
	Product *data.Product
	Place   *data.Place
}

func (s *shopifyProductScorer) GetScore() (int64, error) {
	imgs, err := data.DB.ProductImage.FindByProductID(s.Product.ID)
	if err != nil {
		return 0, err
	}

	var imgScores int64
	//the weight of each individual picture
	pictureWeight := float64(TotalImageWeight) / float64(len(imgs))
	for _, img := range imgs {
		imgScores += img.Score
	}

	//the product score from each image
	score := int64(math.Ceil(pictureWeight * float64(imgScores)))

	if s.Place.IsPriority() {
		score += ScorePriorityMerchant
	}

	if s.Place.Currency != "USD" {
		score += ScoreCurrencyPenalty
	}

	// normalize
	if score < 0 {
		score = 0
	}

	return score, nil
}
