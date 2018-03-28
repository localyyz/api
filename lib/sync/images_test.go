package sync

import (
	"reflect"
	"testing"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
)

type setImagesTest struct {
	name string

	syncImages []*shopify.ProductImage

	fetchProductImages func() ([]*data.ProductImage, error)
	finalize           func(toSaves, toRemoves []*data.ProductImage) error

	expectedSaves   []*data.ProductImage
	expectedRemoves []*data.ProductImage

	t *testing.T // self reference
}

type mockImageSyncer struct {
	*setImagesTest
}

func (m *mockImageSyncer) FetchProductImages() ([]*data.ProductImage, error) {
	if m.fetchProductImages == nil {
		return nil, nil
	}
	return m.fetchProductImages()
}

func (m *mockImageSyncer) Finalize(toSaves, toRemoves []*data.ProductImage) error {
	if m.finalize == nil {
		// compare return and expected values
		if len(toSaves) != len(m.expectedSaves) {
			m.t.Errorf("test '%s': expected toSave length '%d', got '%d'", m.name, len(m.expectedSaves), len(toSaves))
			return nil
		}

		if len(toRemoves) != len(m.expectedRemoves) {
			m.t.Errorf("test '%s': expected toRemoves length '%d', got '%d'", m.name, len(m.expectedRemoves), len(toRemoves))
			return nil
		}

		for i, expected := range m.expectedSaves {
			actual := toSaves[i]
			if !reflect.DeepEqual(expected, actual) {
				m.t.Errorf("test '%s': expected toSave(%d) expected %+v, got %+v", m.name, i, expected, actual)
			}
		}

		for i, expected := range m.expectedRemoves {
			actual := toRemoves[i]
			if !reflect.DeepEqual(expected, actual) {
				m.t.Errorf("test '%s': expected toRemove(%d) expected %+v, got %+v", m.name, i, expected, actual)
			}
		}
		return nil
	}
	return m.finalize(toSaves, toRemoves)
}

func (*mockImageSyncer) GetProduct() *data.Product {
	return &data.Product{ID: 1}
}

func TestSetImages(t *testing.T) {
	t.Parallel()

	tests := []setImagesTest{
		{
			name: "expect no error on no data",
			t:    t,
		},
		{
			name:          "basic first add",
			syncImages:    []*shopify.ProductImage{{ID: 1, Position: 1, Src: "https://link1"}},
			expectedSaves: []*data.ProductImage{{Ordering: 1, ImageURL: "https://link1", ProductID: 1, ExternalID: 1}},
			t:             t,
		},
		{
			name: "keep and do nothing",
			syncImages: []*shopify.ProductImage{
				{ID: 1, Position: 1, Src: "https://link1"},
				{ID: 2, Position: 2, Src: "https://link2"},
			},
			fetchProductImages: func() ([]*data.ProductImage, error) {
				return []*data.ProductImage{
					{ProductID: 1, Ordering: 1, ExternalID: 1, ImageURL: "https://link1"},
					{ProductID: 1, Ordering: 2, ExternalID: 2, ImageURL: "https://link2"},
				}, nil
			},
			t: t,
		},
		{
			name: "keep and create new",
			syncImages: []*shopify.ProductImage{
				{ID: 1, Position: 1, Src: "https://link1"},
				{ID: 2, Position: 2, Src: "https://link2"},
			},
			expectedSaves: []*data.ProductImage{
				{Ordering: 2, ImageURL: "https://link2", ProductID: 1, ExternalID: 2},
			},
			fetchProductImages: func() ([]*data.ProductImage, error) {
				return []*data.ProductImage{
					{ProductID: 1, Ordering: 1, ExternalID: 1, ImageURL: "https://link1"},
				}, nil
			},
			t: t,
		},
		{
			name: "keep and remove one",
			syncImages: []*shopify.ProductImage{
				{ID: 2, Position: 2, Src: "https://link2"},
			},
			expectedRemoves: []*data.ProductImage{
				{Ordering: 1, ImageURL: "https://link1", ProductID: 1, ExternalID: 1},
			},
			fetchProductImages: func() ([]*data.ProductImage, error) {
				return []*data.ProductImage{
					{ProductID: 1, Ordering: 1, ExternalID: 1, ImageURL: "https://link1"},
					{ProductID: 1, Ordering: 2, ExternalID: 2, ImageURL: "https://link2"},
				}, nil
			},
			t: t,
		},
		{
			name: "keep create and remove",
			syncImages: []*shopify.ProductImage{
				{ID: 1, Position: 1, Src: "https://link1"},
				{ID: 3, Position: 2, Src: "https://link3"},
			},
			expectedRemoves: []*data.ProductImage{
				{ExternalID: 2, Ordering: 2, ImageURL: "https://link2", ProductID: 1},
			},
			expectedSaves: []*data.ProductImage{
				{ExternalID: 3, Ordering: 2, ImageURL: "https://link3", ProductID: 1},
			},
			fetchProductImages: func() ([]*data.ProductImage, error) {
				return []*data.ProductImage{
					{ProductID: 1, Ordering: 1, ExternalID: 1, ImageURL: "https://link1"},
					{ProductID: 1, Ordering: 2, ExternalID: 2, ImageURL: "https://link2"},
				}, nil
			},
			t: t,
		},
		{
			name:            "swap (add new one, remove old one)",
			syncImages:      []*shopify.ProductImage{{ID: 2, Position: 1, Src: "https://link2"}},
			expectedSaves:   []*data.ProductImage{{Ordering: 1, ImageURL: "https://link2", ProductID: 1, ExternalID: 2}},
			expectedRemoves: []*data.ProductImage{{Ordering: 1, ImageURL: "https://link1", ProductID: 1, ExternalID: 1}},
			fetchProductImages: func() ([]*data.ProductImage, error) {
				return []*data.ProductImage{
					{ProductID: 1, Ordering: 1, ExternalID: 1, ImageURL: "https://link1"},
				}, nil
			},
			t: t,
		},
		{
			name:          "external variant id is set",
			syncImages:    []*shopify.ProductImage{{ID: 2, Position: 1, Src: "https://link2", VariantIds: []int64{123}}},
			expectedSaves: []*data.ProductImage{{ExternalID: 2, Ordering: 1, ImageURL: "https://link2", ProductID: 1, VariantIDs: []int64{123}}},
			t:             t,
		},
		{
			name: "remove image query params and dedup",
			syncImages: []*shopify.ProductImage{
				{ID: 1, Position: 1, Src: "https://link2?v=1234567"},
				{ID: 1, Position: 1, Src: "https://link2?v=9876543"},
			},
			expectedSaves: []*data.ProductImage{
				{ExternalID: 1, Ordering: 1, ImageURL: "https://link2", ProductID: 1},
			},
			t: t,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			syncer := &mockImageSyncer{&tt}
			setImages(syncer, tt.syncImages...)
		})
	}

}
