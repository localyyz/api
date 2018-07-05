package endtoend

import (
	"fmt"
	"testing"

	"time"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/token"
	"bitbucket.org/moodie-app/moodie-api/web/auth"
	"github.com/go-chi/jwtauth"
	"github.com/stretchr/testify/assert"
)

type fixture struct {
	user1, user2, user3, user4, user5, user6, user7, user8, user9, user10                                                          auth.AuthUser
	testStore                                                                                                                      *data.Place
	productInStock, productNotInStock, lightningProductValid, lightningProductExpired, lightningProductCapHit                      *data.Product
	variantInStock, variantNotInStock, variantWithDiscount, variantLightningValid, variantLightningExpired, variantLightningCapHit *data.ProductVariant
	lightningValid, lightningCapHit, lightningExpired                                                                              *data.Collection
	collectionProductValid, collectionProductExpired, collectionProductCapHit                                                      data.CollectionProduct
	cart                                                                                                                           *data.Cart
	cartItem                                                                                                                       *data.CartItem
	checkout                                                                                                                       *data.Checkout
}

func (f *fixture) setupUser(t *testing.T) {
	f.user1 = newTestUser(t, 1)
	f.user2 = newTestUser(t, 2)
	f.user3 = newTestUser(t, 3)
	f.user4 = newTestUser(t, 4)
	f.user5 = newTestUser(t, 5)
	f.user6 = newTestUser(t, 6)
	f.user7 = newTestUser(t, 7)
	f.user8 = newTestUser(t, 8)
	f.user9 = newTestUser(t, 9)
	f.user10 = newTestUser(t, 10)
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

	f.lightningProductValid = &data.Product{
		Title:   "sample product in lightning collection",
		Status:  data.ProductStatusApproved,
		PlaceID: f.testStore.ID,
	}
	assert.NoError(t, data.DB.Save(f.lightningProductValid))

	f.lightningProductCapHit = &data.Product{
		Title:   "sample product in lightning collection - collection cap hit",
		Status:  data.ProductStatusApproved,
		PlaceID: f.testStore.ID,
	}
	assert.NoError(t, data.DB.Save(f.lightningProductCapHit))

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
	f.variantLightningValid = &data.ProductVariant{
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
	f.variantLightningExpired = &data.ProductVariant{
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
	f.variantLightningCapHit = &data.ProductVariant{
		ProductID: f.lightningProductCapHit.ID,
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
	assert.NoError(t, data.DB.Save(f.variantLightningValid))
	assert.NoError(t, data.DB.Save(f.variantLightningExpired))
	assert.NoError(t, data.DB.Save(f.variantLightningCapHit))
}

func (f *fixture) SetupLightningCollection(t *testing.T) {
	yesterday := time.Now().AddDate(0, 0, -1)
	tomorrow := time.Now().AddDate(0, 0, 1)
	f.lightningValid = &data.Collection{
		Name:        "Valid Collection",
		Description: "Test",
		Lightning:   true,
		StartAt:     &yesterday,
		EndAt:       &tomorrow,
		Cap:         1,
		Status:      data.CollectionStatusActive,
	}
	f.lightningCapHit = &data.Collection{
		Name:        "Cap Hit Collection",
		Description: "Test",
		Lightning:   true,
		StartAt:     &yesterday,
		EndAt:       &tomorrow,
		Cap:         1,
		Status:      data.CollectionStatusActive,
	}
	f.lightningExpired = &data.Collection{
		Name:        "Expired Collection",
		Description: "Test",
		Lightning:   true,
		StartAt:     &yesterday,
		EndAt:       &yesterday,
		Cap:         1,
		Status:      data.CollectionStatusInactive,
	}
	assert.NoError(t, data.DB.Save(f.lightningValid))
	assert.NoError(t, data.DB.Save(f.lightningCapHit))
	assert.NoError(t, data.DB.Save(f.lightningExpired))
}

func (f *fixture) LinkProductsWithCollection(t *testing.T) {

	f.collectionProductValid = data.CollectionProduct{
		CollectionID: f.lightningValid.ID,
		ProductID:    f.lightningProductValid.ID,
	}

	f.collectionProductExpired = data.CollectionProduct{
		CollectionID: f.lightningExpired.ID,
		ProductID:    f.lightningProductExpired.ID,
	}

	f.collectionProductCapHit = data.CollectionProduct{
		CollectionID: f.lightningCapHit.ID,
		ProductID:    f.lightningProductCapHit.ID,
	}

	assert.NoError(t, data.DB.CollectionProduct.Create(f.collectionProductValid))
	assert.NoError(t, data.DB.CollectionProduct.Create(f.collectionProductExpired))
	assert.NoError(t, data.DB.CollectionProduct.Create(f.collectionProductCapHit))
}

func (f *fixture) CreateCart(t *testing.T) {
	f.cart = &data.Cart{
		UserID: f.user9.ID,
		Status: data.CartStatusPaymentSuccess,
	}
	assert.NoError(t, data.DB.Save(f.cart))
}

func (f *fixture) CreateCartItems(t *testing.T) {
	f.cartItem = &data.CartItem{
		CartID:    f.cart.ID,
		ProductID: f.lightningProductCapHit.ID,
		PlaceID:   f.testStore.ID,
		VariantID: f.variantLightningCapHit.ID,
	}
	assert.NoError(t, data.DB.Save(f.cartItem))
}

func (f *fixture) SetupData(t *testing.T) {
	f.setupUser(t)
	f.setupTestStores(t)
	f.setupProduct(t)
	f.SetupLightningCollection(t)
	f.LinkProductsWithCollection(t)
	f.CreateCart(t)
	f.CreateCartItems(t)
}

func (f *fixture) TeardownData(t *testing.T) {
	data.DB.Exec("TRUNCATE users cascade;")
	data.DB.Exec("TRUNCATE places cascade;")
}
