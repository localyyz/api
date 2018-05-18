package sync

import (
	"math"

	"bitbucket.org/moodie-app/moodie-api/data"
	db "upper.io/db.v3"
)

const (
	TOTALIMAGEWEIGHT  = 5
	MINIMUMIMAGEWIDTH = 750
)

func ScoreProduct(product *data.Product) (int64, error) {

	var totalScore int64

	/* getting all the images */
	var images []*data.ProductImage
	res := data.DB.ProductImage.Find(db.Cond{"product_id": product.ID})
	err := res.All(&images)
	if err != nil {
		return 0, err
	}
	totalScore += ScoreImage(images) //score the images

	err = finalizeScore(images) //finalizing(saving) all scores
	if err != nil {
		return 0, err
	}

	return totalScore, nil
}

/*
	A product get a maximum score of 5 from its pictures
	If all its pictures are >= 750 in width its get 5/5
	If not it gets a lower score based on the weighting of each picture
	Each picture is scored out of 1. If width >= 750 it gets 1/1
	Weight of each picture = 5/number of pictures
	score = weight of each pic * number of pics with width >= 750
*/
func ScoreImage(images []*data.ProductImage) int64 {

	if len(images) == 0 {
		return 0
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

	return productScore
}

func finalizeScore(images []*data.ProductImage) error {
	/* save each of the product images */
	for _, val := range images {
		err := data.DB.ProductImage.Save(val)
		if err != nil {
			return err
		}
	}
	return nil
}
