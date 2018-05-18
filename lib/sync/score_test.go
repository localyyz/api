package sync

import (
	"log"
	"testing"

	"bitbucket.org/moodie-app/moodie-api/data"
)

func TestImageScore(t *testing.T) {
	var images [][]*data.ProductImage
	var img1 = []*data.ProductImage{
		{
			ID:     1,
			Width:  750,
			Height: 750,
		},
		{
			ID:     2,
			Width:  800,
			Height: 800,
		},
		{
			ID:     3,
			Width:  850,
			Height: 850,
		},
	}
	var img2 = []*data.ProductImage{
		{
			ID:     4,
			Width:  600,
			Height: 500,
		},
		{
			ID:     5,
			Width:  300,
			Height: 300,
		},
	}
	var img3 = []*data.ProductImage{
		{
			ID:     6,
			Width:  800,
			Height: 800,
		},
		{
			ID:     7,
			Width:  300,
			Height: 300,
		},
		{
			ID:     8,
			Width:  900,
			Height: 900,
		},
	}
	var img4 = []*data.ProductImage{
		{
			ID:     9,
			Width:  300,
			Height: 300,
		},
		{
			ID:     10,
			Width:  300,
			Height: 300,
		},
		{
			ID:     11,
			Width:  800,
			Height: 800,
		},
	}
	var img5 = []*data.ProductImage{
		{
			ID:     12,
			Width:  800,
			Height: 800,
		},
		{
			ID:     13,
			Width:  800,
			Height: 800,
		},
		{
			ID:     14,
			Width:  900,
			Height: 900,
		},
		{
			ID:     15,
			Width:  200,
			Height: 200,
		},
		{
			ID:     16,
			Width:  200,
			Height: 200,
		},
	}
	var img6 []*data.ProductImage
	images = append(images, img1)
	images = append(images, img2)
	images = append(images, img3)
	images = append(images, img4)
	images = append(images, img5)
	images = append(images, img6)

	var expected = []int{5, 0, 4, 2, 3, 0}

	for ind, tt := range expected {
		t.Run("Testing Image Score", func(t *testing.T) {
			imageScore := ScoreImage(images[ind])
			if int64(tt) != imageScore {
				log.Fatalf("Error: Expected Image score %d got %d for test %d ", tt, imageScore, ind)
			}
		})
	}

}
