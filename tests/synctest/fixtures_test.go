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

	testStore *data.Place

	CollectionNotUnique, CollectionUnique *data.Collection
	ProductNotUnique, ProductUnique       *data.Product
}

func (f *fixture) setupProduct(t *testing.T) {
	f.ProductUnique = &data.Product{
		Title:   "sample product 1",
		Status:  data.ProductStatusApproved,
		PlaceID: f.testStore.ID,
	}
	assert.NoError(t, data.DB.Save(f.ProductUnique))

	f.ProductNotUnique = &data.Product{
		Title:   "sample product 2",
		Status:  data.ProductStatusApproved,
		PlaceID: f.testStore.ID,
	}
	assert.NoError(t, data.DB.Save(f.ProductNotUnique))
}

func (f *fixture) setupCollection(t *testing.T) {

	f.CollectionUnique = &data.Collection{
		Name:        "Test Collection 1",
		Description: "Test",
	}

	f.CollectionNotUnique = &data.Collection{
		Name:        "Test Collection 2",
		Description: "Test",
	}
	assert.NoError(t, data.DB.Save(f.CollectionUnique))
	assert.NoError(t, data.DB.Save(f.CollectionNotUnique))
}

func (f *fixture) linkProductsWithCollection(t *testing.T) {

	collectionProduct := data.CollectionProduct{
		CollectionID: f.CollectionNotUnique.ID,
		ProductID:    f.ProductNotUnique.ID,
	}

	assert.NoError(t, data.DB.CollectionProduct.Create(collectionProduct))
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

	f.setupProduct(t)
	f.setupCollection(t)
	f.linkProductsWithCollection(t)
}

func (f *fixture) TeardownData(t *testing.T) {
	data.DB.Exec("TRUNCATE users cascade;")
	data.DB.Exec("TRUNCATE places cascade;")
}
