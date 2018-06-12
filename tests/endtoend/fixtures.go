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
	paul auth.AuthUser

	testStore1 *data.Place
	product1   *data.Product
	variant1   *data.ProductVariant
}

func (f *fixture) setupUser(t *testing.T) {
	// setup fixtures for test suite
	paulUser := &data.User{
		Username:     "paul@localyyz.com",
		Email:        "paul@localyyz.com",
		Name:         "Paul Xue",
		Network:      "email",
		PasswordHash: string(""),
		LoggedIn:     true,
	}
	assert.NoError(t, data.DB.Save(paulUser))
	token, _ := token.Encode(jwtauth.Claims{"user_id": paulUser.ID})
	f.paul = auth.AuthUser{User: paulUser, JWT: token.Raw}
}

func (f *fixture) setupTestStores(t *testing.T) {
	// test stores with actual shopify cred
	f.testStore1 = &data.Place{Name: "best merchant"}
	assert.NoError(t, data.DB.Save(f.testStore1))
	assert.NoError(t, data.DB.Save(&data.ShopifyCred{
		PlaceID:     f.testStore1.ID,
		AccessToken: "79314e78d313b5c4d2cd9dcd1121d0dd",
		ApiURL:      "https://localyyz-dev-shop.myshopify.com",
	}))
}

func (f *fixture) setupProduct(t *testing.T) {
	// product
	f.product1 = &data.Product{
		Title:   "best product",
		Status:  data.ProductStatusApproved,
		PlaceID: f.testStore1.ID,
	}
	assert.NoError(t, data.DB.Save(f.product1))

	// product variant
	f.variant1 = &data.ProductVariant{
		ProductID: f.product1.ID,
		PlaceID:   f.testStore1.ID,
		Price:     10,
		Limits:    10,
		OfferID:   32226754053,
	}
	data.DB.Save(f.variant1)
}

func (f *fixture) SetupData(t *testing.T) {
	f.setupUser(t)
	f.setupTestStores(t)
	f.setupProduct(t)
}
