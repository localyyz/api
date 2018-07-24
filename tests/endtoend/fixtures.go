package endtoend

import (
	"context"
	"fmt"
	"testing"

	"time"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/tests/apiclient"
	"bitbucket.org/moodie-app/moodie-api/web/auth"
	"github.com/stretchr/testify/assert"
)

type UserClient struct {
	*auth.AuthUser
	client *apiclient.Client
}

type fixture struct {
	apiURL string

	user, user2 *UserClient
	testStore   *data.Place

	productInStock, productNotInStock                      *data.Product
	variantInStock, variantNotInStock, variantWithDiscount *data.ProductVariant

	lightningProductValid, lightningProductExpired *data.Product
	variantDealValid, variantDealExpired           *data.ProductVariant
	dealValid, dealExpired                         *data.Collection
}

func (f *fixture) setupUser(t *testing.T) {
	f.user = f.newTestUser(t, 1)
	f.user2 = f.newTestUser(t, 2)

}

func (f *fixture) newTestUser(t *testing.T, n int) *UserClient {
	client, err := apiclient.NewClient(f.apiURL)
	assert.NoError(t, err)

	// setup fixtures for test suite
	ctx := context.Background()
	authUser, _, err := client.User.Signup(
		ctx,
		&data.User{
			Name:  "Paul X",
			Email: fmt.Sprintf("test%d@localyyz.com", n),
		},
	)
	assert.NoError(t, err)

	// setup clients JWT
	client.JWT(authUser.JWT)

	return &UserClient{
		AuthUser: authUser,
		client:   client,
	}
}

func (f *fixture) setupTestStores(t *testing.T) {
	// test stores with actual shopify cred
	f.testStore = &data.Place{Name: "best merchant"}
	assert.NoError(t, data.DB.Save(f.testStore))
	assert.NoError(t, data.DB.Save(&data.ShopifyCred{
		PlaceID:     f.testStore.ID,
		AccessToken: "ab4f6c5f522de90702e52f95e3e72d88",
		ApiURL:      "https://best-test-store-toronto.myshopify.com",
	}))
}

func (f *fixture) setupProduct(t *testing.T) {
	f.productInStock = &data.Product{
		Title:   "sample product",
		Status:  data.ProductStatusApproved,
		PlaceID: f.testStore.ID,
	}
	assert.NoError(t, data.DB.Save(f.productInStock))

	f.productNotInStock = &data.Product{
		Title:   "sample product not in stock",
		Status:  data.ProductStatusApproved,
		PlaceID: f.testStore.ID,
	}
	assert.NoError(t, data.DB.Save(f.productNotInStock))

	// lightning
	f.lightningProductValid = &data.Product{
		Title:   "sample product in lightning collection",
		Status:  data.ProductStatusApproved,
		PlaceID: f.testStore.ID,
	}
	assert.NoError(t, data.DB.Save(f.lightningProductValid))

	f.lightningProductExpired = &data.Product{
		Title:   "sample product in lightning collection - collection expired",
		Status:  data.ProductStatusApproved,
		PlaceID: f.testStore.ID,
	}
	assert.NoError(t, data.DB.Save(f.lightningProductExpired))

	// NOTE: https://best-test-store-toronto.myshopify.com/admin/products/10761547971.json
	f.variantInStock = &data.ProductVariant{
		ProductID: f.productInStock.ID,
		PlaceID:   f.testStore.ID,
		Price:     10,
		Limits:    15,
		OfferID:   43252300547,
		Etc: data.ProductVariantEtc{
			Size:  "small",
			Color: "deep",
		},
	}
	f.variantNotInStock = &data.ProductVariant{
		ProductID: f.productNotInStock.ID,
		PlaceID:   f.testStore.ID,
		Price:     10,
		Limits:    10,
		OfferID:   43252300483,
	}
	f.variantWithDiscount = &data.ProductVariant{
		ProductID: f.productInStock.ID,
		PlaceID:   f.testStore.ID,
		Price:     10,
		Limits:    15,
		OfferID:   43252300611,
	}
	f.variantDealValid = &data.ProductVariant{
		ProductID: f.lightningProductValid.ID,
		PlaceID:   f.testStore.ID,
		Price:     10,
		Limits:    10,
		OfferID:   43252300547,
		Etc: data.ProductVariantEtc{
			Size:  "small",
			Color: "deep",
		},
	}
	f.variantDealExpired = &data.ProductVariant{
		ProductID: f.lightningProductExpired.ID,
		PlaceID:   f.testStore.ID,
		Price:     10,
		Limits:    10,
		OfferID:   43252300547,
		Etc: data.ProductVariantEtc{
			Size:  "small",
			Color: "deep",
		},
	}

	assert.NoError(t, data.DB.Save(f.variantInStock))
	assert.NoError(t, data.DB.Save(f.variantNotInStock))
	assert.NoError(t, data.DB.Save(f.variantWithDiscount))
	assert.NoError(t, data.DB.Save(f.variantDealValid))
	assert.NoError(t, data.DB.Save(f.variantDealExpired))
}

func (f *fixture) setupDealCollection(t *testing.T) {
	yesterday := time.Now().AddDate(0, 0, -1)
	tomorrow := time.Now().AddDate(0, 0, 1)
	f.dealValid = &data.Collection{
		Name:        "Valid Collection",
		Description: "Test",
		Lightning:   true,
		StartAt:     &yesterday,
		EndAt:       &tomorrow,
		Cap:         1,
		Status:      data.CollectionStatusActive,
	}

	f.dealExpired = &data.Collection{
		Name:        "Expired Collection",
		Description: "Test",
		Lightning:   true,
		StartAt:     &yesterday,
		EndAt:       &yesterday,
		Cap:         1,
		Status:      data.CollectionStatusInactive,
	}
	assert.NoError(t, data.DB.Save(f.dealValid))
	assert.NoError(t, data.DB.Save(f.dealExpired))
}

func (f *fixture) linkProductsWithCollection(t *testing.T) {
	collectionProductValid := data.CollectionProduct{
		CollectionID: f.dealValid.ID,
		ProductID:    f.lightningProductValid.ID,
	}

	collectionProductExpired := data.CollectionProduct{
		CollectionID: f.dealExpired.ID,
		ProductID:    f.lightningProductExpired.ID,
	}

	assert.NoError(t, data.DB.CollectionProduct.Create(collectionProductValid))
	assert.NoError(t, data.DB.CollectionProduct.Create(collectionProductExpired))
}

func (f *fixture) SetupData(t *testing.T, apiURL string) {
	f.apiURL = apiURL

	f.setupUser(t)
	f.setupTestStores(t)
	f.setupProduct(t)
	f.setupDealCollection(t)
	f.linkProductsWithCollection(t)
}

func (f *fixture) TeardownData(t *testing.T) {
	data.DB.Exec("TRUNCATE users cascade;")
	data.DB.Exec("TRUNCATE places cascade;")
}

type MockFacebook struct{}

func (f *MockFacebook) Login(token, inviteCode string) (*data.User, error) {
	if token == "localyyz-test-token-login" {
		user := data.User{ID: 0, Network: "facebook", Email: "test@localyyz.com"}
		return &user, nil
	}
	return nil, nil
}

func (f *MockFacebook) GetUser(u *data.User) error {
	return nil
}
