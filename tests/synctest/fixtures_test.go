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

	// out of stock listing
	outOfStock []*shopify.ProductList

	// discount
	discount []*shopify.ProductList

	// other tests, see comments
	parse1 []*shopify.ProductList

	testStore       *data.Place
	testStoreFemale *data.Place

	// collection sync test
	collection *data.Collection
	product    *data.Product
}

func (f *fixture) setupProduct(t *testing.T) {
	f.product = &data.Product{
		Title:   "sample product 1",
		Status:  data.ProductStatusApproved,
		PlaceID: f.testStore.ID,
	}
	assert.NoError(t, data.DB.Save(f.product))
}

func (f *fixture) setupCollection(t *testing.T) {

	f.collection = &data.Collection{
		Name:        "Test Collection 1",
		Description: "Test",
	}
	assert.NoError(t, data.DB.Save(f.collection))
}

func (f *fixture) SetupData(t *testing.T) {
	type wrapper struct {
		ProductListings []*shopify.ProductList `json:"product_listings"`
	}

	{ // initial create
		w := &wrapper{}
		assert.NoError(t, json.Unmarshal(fixtures.ProductListingsCreate, &w))
		f.createListing = w.ProductListings
		assert.NotEmpty(t, f.createListing)
	}

	{ // bump inventory by 10
		w := &wrapper{}
		assert.NoError(t, json.Unmarshal(fixtures.ProductListingsAddInventory, &w))
		f.updateListing = w.ProductListings
		assert.NotEmpty(t, f.updateListing)
	}

	{ // swap first image
		w := &wrapper{}
		assert.NoError(t, json.Unmarshal(fixtures.ProductListingsSwapImage, &w))
		f.swapImageUpdate = w.ProductListings
		assert.NotEmpty(t, f.swapImageUpdate)
	}

	{ // out of stock
		w := &wrapper{}
		assert.NoError(t, json.Unmarshal(fixtures.ProductListingsOutOfStock, &w))
		f.outOfStock = w.ProductListings
		assert.NotEmpty(t, f.outOfStock)
	}

	{ // discount
		w := &wrapper{}
		assert.NoError(t, json.Unmarshal(fixtures.ProductListingsDiscount, &w))
		f.discount = w.ProductListings
		assert.NotEmpty(t, f.discount)
	}

	{ // product list parse testing 1 -> correct gender parsed from description
		w := &wrapper{}
		assert.NoError(t, json.Unmarshal(fixtures.ProductListingsParse1, &w))
		f.parse1 = w.ProductListings
		assert.NotEmpty(t, f.parse1)
	}

	// setup shop
	f.testStore = &data.Place{
		Name:   "best merchant",
		Gender: data.PlaceGenderUnisex,
		Weight: int32(1),
	}
	assert.NoError(t, data.DB.Save(f.testStore))

	// setup shop
	f.testStoreFemale = &data.Place{
		Name:   "best merchant female",
		Gender: data.PlaceGenderFemale,
		Weight: int32(1),
	}
	assert.NoError(t, data.DB.Save(f.testStoreFemale))

	malePants := &data.Category{
		ID:    1001,
		Label: "Pants",
		Value: "pants",
	}
	femalePants := &data.Category{
		ID:    2001,
		Label: "Pants",
		Value: "pants",
	}
	assert.NoError(t, data.DB.Category.Create(malePants))
	assert.NoError(t, data.DB.Category.Create(femalePants))

	// setup category
	assert.NoError(t, data.DB.Whitelist.Create(data.Whitelist{
		Type:       data.CategoryApparel,
		Value:      "pants",
		Gender:     data.ProductGenderMale,
		Weight:     1,
		CategoryID: &(malePants.ID),
	}))
	assert.NoError(t, data.DB.Whitelist.Create(data.Whitelist{
		Type:       data.CategoryApparel,
		Value:      "pants",
		Gender:     data.ProductGenderFemale,
		Weight:     1,
		CategoryID: &(femalePants.ID),
	}))

	f.setupProduct(t)
	f.setupCollection(t)
}

func (f *fixture) TeardownData(t *testing.T) {
	data.DB.Exec("TRUNCATE users cascade;")
	data.DB.Exec("TRUNCATE places cascade;")
	data.DB.Exec("TRUNCATE collections cascade;")
	data.DB.Exec("TRUNCATE categories cascade;")
	data.DB.Exec("TRUNCATE whitelist cascade;")
}
