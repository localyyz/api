package endtoend

import (
	"testing"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/token"
	"bitbucket.org/moodie-app/moodie-api/web/auth"
	"github.com/go-chi/jwtauth"
	"github.com/stretchr/testify/assert"
)

type fixture struct {
	user1, user2, user3, user4, user5, user6               auth.AuthUser
	testStore                                              *data.Place
	productInStock, productNotInStock                      *data.Product
	variantInStock, variantNotInStock, variantWithDiscount *data.ProductVariant
}

func (f *fixture) setupUser(t *testing.T) {
	// setup fixtures for test suite
	user1 := &data.User{
		Username:     "waseef@localyyz.com",
		Email:        "waseef@localyyz.com",
		Name:         "Waseef Shawkat",
		Network:      "email",
		PasswordHash: string(""),
		LoggedIn:     true,
	}
	assert.NoError(t, data.DB.Save(user1))
	token1, _ := token.Encode(jwtauth.Claims{"user_id": user1.ID})
	f.user1 = auth.AuthUser{User: user1, JWT: token1.Raw}

	user2 := &data.User{
		Username:     "waseef2@localyyz.com",
		Email:        "waseef@localyyz.com",
		Name:         "Paul Xue",
		Network:      "email",
		PasswordHash: string(""),
		LoggedIn:     true,
	}
	assert.NoError(t, data.DB.Save(user2))
	token2, _ := token.Encode(jwtauth.Claims{"user_id": user2.ID})
	f.user2 = auth.AuthUser{User: user2, JWT: token2.Raw}

	user3 := &data.User{
		Username:     "waseef3@localyyz.com",
		Email:        "waseef2@localyyz.com",
		Name:         "Paul Xue",
		Network:      "email",
		PasswordHash: string(""),
		LoggedIn:     true,
	}
	assert.NoError(t, data.DB.Save(user3))
	token3, _ := token.Encode(jwtauth.Claims{"user_id": user3.ID})
	f.user3 = auth.AuthUser{User: user3, JWT: token3.Raw}

	// setup fixtures for test suite
	user4 := &data.User{
		Username:     "waseef4@localyyz.com",
		Email:        "waseef@localyyz.com",
		Name:         "Waseef Shawkat",
		Network:      "email",
		PasswordHash: string(""),
		LoggedIn:     true,
	}
	assert.NoError(t, data.DB.Save(user4))
	token4, _ := token.Encode(jwtauth.Claims{"user_id": user4.ID})
	f.user4 = auth.AuthUser{User: user4, JWT: token4.Raw}

	// setup fixtures for test suite
	user5 := &data.User{
		Username:     "waseef5@localyyz.com",
		Email:        "waseef@localyyz.com",
		Name:         "Waseef Shawkat",
		Network:      "email",
		PasswordHash: string(""),
		LoggedIn:     true,
	}
	assert.NoError(t, data.DB.Save(user5))
	token5, _ := token.Encode(jwtauth.Claims{"user_id": user5.ID})
	f.user5 = auth.AuthUser{User: user5, JWT: token5.Raw}

	// setup fixtures for test suite
	user6 := &data.User{
		Username:     "paul6@localyyz.com",
		Email:        "paul@localyyz.com",
		Name:         "Paul X",
		Network:      "email",
		PasswordHash: string(""),
		LoggedIn:     true,
	}
	assert.NoError(t, data.DB.Save(user6))
	token6, _ := token.Encode(jwtauth.Claims{"user_id": user6.ID})
	f.user6 = auth.AuthUser{User: user6, JWT: token6.Raw}
}

func (f *fixture) setupTestStores(t *testing.T) {
	// test stores with actual shopify cred
	f.testStore = &data.Place{Name: "best merchant"}
	assert.NoError(t, data.DB.Save(f.testStore))
	assert.NoError(t, data.DB.Save(&data.ShopifyCred{
		PlaceID:     f.testStore.ID,
		AccessToken: "79314e78d313b5c4d2cd9dcd1121d0dd",
		ApiURL:      "https://localyyz-dev-shop.myshopify.com",
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

	f.variantInStock = &data.ProductVariant{
		ProductID: f.productInStock.ID,
		PlaceID:   f.testStore.ID,
		Price:     10,
		Limits:    10,
		OfferID:   32226754053,
	}
	f.variantNotInStock = &data.ProductVariant{
		ProductID: f.productNotInStock.ID,
		PlaceID:   f.testStore.ID,
		Price:     10,
		Limits:    10,
		OfferID:   32294113669,
	}
	f.variantWithDiscount = &data.ProductVariant{
		ProductID: f.productInStock.ID,
		PlaceID:   f.testStore.ID,
		Price:     10,
		Limits:    10,
		OfferID:   32226754501,
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
