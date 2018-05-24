package sync

import (
	"log"
	"testing"

	"bitbucket.org/moodie-app/moodie-api/data"
)

type mockImageScorer struct {
	Place   *data.Place
	Product *data.Product
}

func (m *mockImageScorer) GetProductImages(ID int64) ([]*data.ProductImage, error) {
	if ID == 1 || ID == 6 {
		return []*data.ProductImage{
			{Width: 750, Height: 750},
			{Width: 800, Height: 800},
			{Width: 850, Height: 850},
		}, nil
	} else if ID == 2 {
		return []*data.ProductImage{
			{Width: 600, Height: 500},
			{Width: 300, Height: 300},
		}, nil
	} else if ID == 3 {
		return []*data.ProductImage{
			{Width: 800, Height: 800},
			{Width: 300, Height: 300},
			{Width: 900, Height: 900},
		}, nil
	} else if ID == 4 {
		return []*data.ProductImage{
			{Width: 300, Height: 300},
			{Width: 300, Height: 300},
			{Width: 800, Height: 800},
		}, nil
	} else if ID == 5 {
		return []*data.ProductImage{}, nil
	} else {
		return nil, nil
	}
}

func (m *mockImageScorer) GetProduct() *data.Product {
	return m.Product
}

func (m *mockImageScorer) GetPlace() *data.Place {
	return m.Place
}

func (m *mockImageScorer) CheckPriority() bool {
	if m.GetPlace().ID == 1 {
		return true
	} else {
		return false
	}
}

func (m *mockImageScorer) Finalize(imgs []*data.ProductImage) error {
	return nil
}

type testInput struct {
	Place   *data.Place
	Product *data.Product
}

type test struct {
	name     string
	input    *testInput
	expected int64
}

func TestImageScore(t *testing.T) {
	tests := []struct {
		name     string
		input    []*data.ProductImage
		expected []int64
	}{
		{"All > 750", []*data.ProductImage{{Width: 800, Height: 800}, {Width: 900, Height: 900}}, []int64{1, 1}},
		{"All < 750", []*data.ProductImage{{Width: 200, Height: 200}, {Width: 200, Height: 200}}, []int64{0, 0}},
		{"1 out of 2 > 750", []*data.ProductImage{{Width: 800, Height: 800}, {Width: 100, Height: 100}}, []int64{1, 0}},
		{"empty list", []*data.ProductImage{}, []int64{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scoreIndividualImages(tt.input)
			for ind, val := range tt.input {
				if val.Score != tt.expected[ind] {
					log.Fatalf("Error: Expected Image score %d got %d for test %s", tt.expected[ind], val.Score, tt.name)
				}
			}
		})
	}
}

func TestProductScore(t *testing.T) {

	tests := []struct {
		name     string
		test     *testInput
		expected int64
	}{
		{"all images width > 750", &testInput{Place: &data.Place{Currency: "USD"}, Product: &data.Product{ID: 1}}, 5},
		{"all images width > 750 and priority merchant", &testInput{Place: &data.Place{Currency: "USD", ID: 1}, Product: &data.Product{ID: 1}}, 6},
		{"all images width < 750", &testInput{Place: &data.Place{Currency: "USD"}, Product: &data.Product{ID: 2}}, 0},
		{"2 out of 3 images width > 750", &testInput{Place: &data.Place{Currency: "USD"}, Product: &data.Product{ID: 3}}, 4},
		{"2 out of 3 images width < 750", &testInput{Place: &data.Place{Currency: "USD"}, Product: &data.Product{ID: 4}}, 2},
		{"empty list of images", &testInput{Place: &data.Place{Currency: "USD"}, Product: &data.Product{ID: 5}}, 0},
		{"all images width > 750 but not USD", &testInput{Place: &data.Place{Currency: "CAD"}, Product: &data.Product{ID: 1}}, 4},
		{"all images width > 750 but not USD and priority merchant", &testInput{Place: &data.Place{Currency: "CAD", ID: 1}, Product: &data.Product{ID: 1}}, 5},
		{"all images width < 750 but not USD", &testInput{Place: &data.Place{Currency: "CAD"}, Product: &data.Product{ID: 2}}, -1},
		{"2 out of 3 images width > 750 but not USD", &testInput{Place: &data.Place{Currency: "CAD"}, Product: &data.Product{ID: 3}}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scoreProduct(&mockImageScorer{Place: tt.test.Place, Product: tt.test.Product})
			if tt.expected != tt.test.Product.Score {
				log.Fatalf("Error: Expected Product score %d got %d for test %s", tt.expected, tt.test.Product.Score, tt.name)
			}
		})
	}
}
