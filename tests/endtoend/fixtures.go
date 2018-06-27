package endtoend

import (
	"fmt"
	"testing"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/token"
	"bitbucket.org/moodie-app/moodie-api/web/auth"
	"github.com/go-chi/jwtauth"
	"github.com/stretchr/testify/assert"
)

type fixture struct {
	user1, user2, user3, user4, user5, user6, user7        auth.AuthUser
	testStore                                              *data.Place
	productInStock, productNotInStock                      *data.Product
	variantInStock, variantNotInStock, variantWithDiscount *data.ProductVariant
}

func (f *fixture) setupUser(t *testing.T) {
	f.user1 = newTestUser(t, 1)
	f.user2 = newTestUser(t, 2)
	f.user3 = newTestUser(t, 3)
	f.user4 = newTestUser(t, 4)
	f.user5 = newTestUser(t, 5)
	f.user6 = newTestUser(t, 6)
	f.user7 = newTestUser(t, 7)
}

func newTestUser(t *testing.T, n int) auth.AuthUser {
	// setup fixtures for test suite
	user := &data.User{
		Username:     fmt.Sprintf("user%d", n),
		Email:        "paul@localyyz.com",
		Name:         "Paul X",
		Network:      "email",
		PasswordHash: string(""),
		LoggedIn:     true,
	}
	assert.NoError(t, data.DB.Save(user))
	token, _ := token.Encode(jwtauth.Claims{"user_id": user.ID})
	return auth.AuthUser{User: user, JWT: token.Raw}
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

	// NOTE: https://best-test-store-toronto.myshopify.com/admin/products/10761547971.json
	f.variantInStock = &data.ProductVariant{
		ProductID: f.productInStock.ID,
		PlaceID:   f.testStore.ID,
		Price:     10,
		Limits:    10,
		OfferID:   43252300547,
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
		Limits:    10,
		OfferID:   43252300611,
	}
	assert.NoError(t, data.DB.Save(f.variantInStock))
	assert.NoError(t, data.DB.Save(f.variantNotInStock))
	assert.NoError(t, data.DB.Save(f.variantWithDiscount))
}

func (f *fixture) SetupData(t *testing.T) {
	f.setupUser(t)
	f.setupTestStores(t)
	f.setupProduct(t)
}

func (f *fixture) TeardownData(t *testing.T) {
	data.DB.Exec("TRUNCATE users cascade;")
	data.DB.Exec("TRUNCATE places cascade;")
}
