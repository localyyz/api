package sync

import (
	"log"
	"math"
	"net/http"
	"testing"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/pkg/errors"
)

type testImageScorer struct {
}

func (m *testImageScorer) ScoreProductImages(images []*data.ProductImage) error {

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

func (m *testImageScorer) Finalize(images []*data.ProductImage) int64 {

	if len(images) == 0 {
		return 0
	}

	var totalScore int64
	totalScore = 0
	var productScore int64
	productScore = 0

	pictureWeight := float64(TotalImageWeight) / float64(len(images)) //the weight of each individual picture
	for _, img := range images {
		if img.Score == 1 {
			totalScore++
		}
	}
	productScore = int64(math.Ceil(pictureWeight * float64(totalScore))) //the product score from each image
	return productScore
}

func TestImageScore(t *testing.T) {
	imageScorer := &testImageScorer{}
	var images [][]*data.ProductImage
	var img1 = []*data.ProductImage{
		{
			ID:       1,
			Width:    750,
			Height:   750,
			ImageURL: "http://www.google.com",
		},
		{
			ID:       2,
			Width:    800,
			Height:   800,
			ImageURL: "http://www.google.com",
		},
		{
			ID:       3,
			Width:    850,
			Height:   850,
			ImageURL: "http://www.google.com",
		},
	}
	var img2 = []*data.ProductImage{
		{
			ID:       4,
			Width:    600,
			Height:   500,
			ImageURL: "http://www.google.com",
		},
		{
			ID:       5,
			Width:    300,
			Height:   300,
			ImageURL: "http://www.google.com",
		},
	}
	var img3 = []*data.ProductImage{
		{
			ID:       6,
			Width:    800,
			Height:   800,
			ImageURL: "http://www.google.com",
		},
		{
			ID:       7,
			Width:    300,
			Height:   300,
			ImageURL: "http://www.google.com",
		},
		{
			ID:       8,
			Width:    900,
			Height:   900,
			ImageURL: "http://www.google.com",
		},
	}
	var img4 = []*data.ProductImage{
		{
			ID:       9,
			Width:    300,
			Height:   300,
			ImageURL: "http://www.google.com",
		},
		{
			ID:       10,
			Width:    300,
			Height:   300,
			ImageURL: "http://www.google.com",
		},
		{
			ID:       11,
			Width:    800,
			Height:   800,
			ImageURL: "http://www.google.com",
		},
	}
	var img5 = []*data.ProductImage{
		{
			ID:       12,
			Width:    800,
			Height:   800,
			ImageURL: "http://www.google.com",
		},
		{
			ID:       13,
			Width:    800,
			Height:   800,
			ImageURL: "http://www.google.com",
		},
		{
			ID:       14,
			Width:    900,
			Height:   900,
			ImageURL: "http://www.google.com",
		},
		{
			ID:       15,
			Width:    200,
			Height:   200,
			ImageURL: "http://www.google.com",
		},
		{
			ID:       16,
			Width:    200,
			Height:   200,
			ImageURL: "http://www.google.com",
		},
	}
	var img6 = []*data.ProductImage{
		{
			ID:       17,
			Width:    900,
			Height:   900,
			ImageURL: "https://cdn.shopify.com/s/files/1/2380/1137/products/21ee6429441e966a7b6801af2f2b0af4.jpg",
		},
	}
	var img7 []*data.ProductImage

	images = append(images, img1)
	images = append(images, img2)
	images = append(images, img3)
	images = append(images, img4)
	images = append(images, img5)
	images = append(images, img6)
	images = append(images, img7)
	var expected = []int{5, 0, 4, 2, 3, 0, 0}

	for ind, tt := range expected {
		t.Run("Testing Image Score", func(t *testing.T) {
			imageScorer.ScoreProductImages(images[ind])
			imageScore := imageScorer.Finalize(images[ind])
			if int64(tt) != imageScore {
				log.Fatalf("Error: Expected Image score %d got %d for test %d ", tt, imageScore, ind)
			}
		})
	}

}
