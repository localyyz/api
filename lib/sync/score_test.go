package sync

import (
	"log"
	"testing"

	"bitbucket.org/moodie-app/moodie-api/data"
)

func TestImageScore(t *testing.T) {
	var images [][]*data.ProductImage
	var products = []*data.Product{
		{
			ID: 1,
		},
		{
			ID: 2,
		},
		{
			ID: 3,
		},
		{
			ID: 4,
		},
		{
			ID: 5,
		},
		{
			ID: 6,
		},
		{
			ID: 7,
		},
	}
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
			ImageURL: "https://404",
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
			imageScorer := &shopifyImageScorer{Product: products[ind], Client: &MockClient{}}
			imageScorer.ScoreProductImages(images[ind])
			imageScorer.Finalize(images[ind])
			if int64(tt) != imageScorer.GetProduct().Score {
				log.Fatalf("Error: Expected Image score %d got %d for test %d ", tt, imageScorer.GetProduct().Score, ind)
			}
		})
	}

}
