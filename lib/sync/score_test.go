package sync

import (
	"log"
	"testing"

	"bitbucket.org/moodie-app/moodie-api/data"
)

type scoreTest struct {
	name     string
	input    []*data.ProductImage
	expected int
}

func TestImageScore(t *testing.T) {
	var tests = []scoreTest{
		{
			name: "All Imgs Width > 750",
			input: []*data.ProductImage{
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
			},
			expected: 5,
		},
		{
			name: "All Imgs Width < 750",
			input: []*data.ProductImage{
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
			},
			expected: 0,
		},
		{
			name: "2 out of 3 images width > 750",
			input: []*data.ProductImage{
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
			},
			expected: 4,
		},
		{
			name: "2 out of 3 images width < 750",
			input: []*data.ProductImage{
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
			},
			expected: 2,
		},
		{
			name: "3 out of 5 images width > 750",
			input: []*data.ProductImage{
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
			},
			expected: 3,
		},
		{
			name: "Invalid Image url",
			input: []*data.ProductImage{
				{
					ID:       17,
					Width:    900,
					Height:   900,
					ImageURL: "https://404",
				},
			},
			expected: 0,
		},
		{
			name:     "Empty List of Images",
			input:    []*data.ProductImage{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product := &data.Product{}
			imageScorer := &shopifyImageScorer{Product: product, Client: &MockClient{}}
			imageScorer.ScoreProductImages(tt.input)
			imageScorer.Finalize(tt.input)
			if int64(tt.expected) != imageScorer.GetProduct().Score {
				log.Fatalf("Error: Expected Image score %d got %d for test %d ", tt.expected, imageScorer.GetProduct().Score, tt.expected)
			}
		})
	}

}
