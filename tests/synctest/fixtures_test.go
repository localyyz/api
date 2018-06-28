package synctest

import (
	"encoding/json"
	"testing"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/tests/synctest/fixtures"
	"github.com/stretchr/testify/assert"
)

type fixture struct {
	createListing []*shopify.ProductList

	// bump all inventory by 10
	updateListing []*shopify.ProductList

	// remove one image and add a new one
	swapImageUpdate []*shopify.ProductList

	testStore *data.Place
}

func (f *fixture) SetupData(t *testing.T) {
	wrapper := struct {
		ProductListings []*shopify.ProductList `json:"product_listings"`
	}{}
	assert.NoError(t, json.Unmarshal(fixtures.ProductListingsCreate, &wrapper))

	f.createListing = wrapper.ProductListings
	assert.NotEmpty(t, f.createListing)

	for _, p := range f.createListing {
		pp := *p
		// update a few things
		for _, pv := range pp.Variants {
			pv.InventoryQuantity += 10
		}
		f.updateListing = append(f.updateListing, &pp)
	}

	for _, p := range f.createListing {
		pp := *p
		// splice the first indexed image
		swap := []*shopify.ProductImage{
			{
				ID:       123123123,
				Position: 1,
				Width:    1500,
				Height:   1500,
				Src:      "https://cdn.shopify.com/s/files/1/1976/6885/files/01-Lit-SpringSummer-Sale-App-Banner.jpg?6158053899319648995",
			},
		}
		pp.Images = append(swap, pp.Images[1:]...)
		f.swapImageUpdate = append(f.swapImageUpdate, &pp)
	}

	for _, p := range f.createListing {
		pp := *p
		// splice the first indexed image
		swap := []*shopify.ProductImage{
			{
				ID:       123123123,
				Position: 1,
				Width:    1500,
				Height:   1500,
				Src:      "https://cdn.shopify.com/s/files/1/1976/6885/files/01-Lit-SpringSummer-Sale-App-Banner.jpg?6158053899319648995",
			},
		}
		pp.Images = append(swap, pp.Images[1:]...)
		f.swapImageUpdate = append(f.swapImageUpdate, &pp)
	}

	// setup shop
	f.testStore = &data.Place{
		Name:   "best merchant",
		Gender: data.PlaceGenderUnisex,
		Weight: int32(1),
	}
	assert.NoError(t, data.DB.Save(f.testStore))

	// setup category
	assert.NoError(t, data.DB.Category.Create(data.Category{
		Type:    data.CategoryApparel,
		Value:   "pants",
		Mapping: "pants",
		Gender:  data.ProductGenderUnisex,
		Weight:  1,
	}))
}

func (f *fixture) TeardownData(t *testing.T) {
	data.DB.Exec("TRUNCATE users cascade;")
	data.DB.Exec("TRUNCATE places cascade;")
}
