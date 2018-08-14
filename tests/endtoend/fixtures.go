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

	user, user2   *UserClient
	anonUser      *UserClient
	testStore     *data.Place
	testStoreCred *data.ShopifyCred

	// address
	validAddress *data.CartAddress

	// products and variants
	productInStock, productNotInStock *data.Product

	variantInStock, variantNotInStock, variantNotInStockRemotely *data.ProductVariant

	// discount
	variantWithDiscount *data.ProductVariant

	// deals
	dealActive        *data.Deal
	productDealActive *data.Product
	variantDealActive *data.ProductVariant

	dealExpired        *data.Deal
	productDealExpired *data.Product
	variantDealExpired *data.ProductVariant
}

func (f *fixture) setupUser(t *testing.T) {
	f.user = f.newTestUser(t, 1)
	f.user2 = f.newTestUser(t, 2)
	f.anonUser = f.newAnonUser(t, "1")
}

func (f *fixture) newAnonUser(t *testing.T, count string) *UserClient {
	client, err := apiclient.NewClient(f.apiURL)
	assert.NoError(t, err)

	mockDeviceId := "localyyz_device_user_" + count
	// this is used to make api calls with device id
	client.AddHeader("X-DEVICE-ID", mockDeviceId)

	// ping the api. should create a mock device user
	_, _, err = client.Cart.Get(context.Background())
	assert.NoError(t, err)

	// we need to read in the ID from the DB
	mockUser, _ := data.DB.User.FindByUsername(mockDeviceId)

	return &UserClient{
		AuthUser: &auth.AuthUser{
			User: &data.User{
				ID:       mockUser.ID,
				Username: mockDeviceId,
				Email:    mockDeviceId,
			},
		},
		client: client,
	}
}

func (f *fixture) newTestUser(t *testing.T, n int) *UserClient {
	client, err := apiclient.NewClient(f.apiURL)
	assert.NoError(t, err)

	// setup fixtures for test suite
	ctx := context.Background()
	authUser, _, err := client.User.SignupWithEmail(
		ctx,
		fmt.Sprintf("test%d@localyyz.com", n),
		"Test Localyyz",
		"test1234",
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
	f.testStoreCred = &data.ShopifyCred{
		PlaceID:     f.testStore.ID,
		AccessToken: "ab4f6c5f522de90702e52f95e3e72d88",
		ApiURL:      "https://best-test-store-toronto.myshopify.com",
	}
	assert.NoError(t, data.DB.Save(f.testStoreCred))
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
	f.productDealActive = &data.Product{
		Title:   "product with active deal",
		Status:  data.ProductStatusApproved,
		PlaceID: f.testStore.ID,
	}
	f.productDealExpired = &data.Product{
		Title:   "product with expired deal",
		Status:  data.ProductStatusApproved,
		PlaceID: f.testStore.ID,
	}
	assert.NoError(t, data.DB.Save(f.productNotInStock))
	assert.NoError(t, data.DB.Save(f.productDealActive))
	assert.NoError(t, data.DB.Save(f.productDealExpired))

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
		Limits:    0,
		OfferID:   43252300483,
	}
	f.variantNotInStockRemotely = &data.ProductVariant{
		ProductID: f.productNotInStock.ID,
		PlaceID:   f.testStore.ID,
		Price:     10,
		Limits:    10,
		OfferID:   12482519760951,
	}
	f.variantWithDiscount = &data.ProductVariant{
		ProductID: f.productInStock.ID,
		PlaceID:   f.testStore.ID,
		Price:     10,
		Limits:    15,
		OfferID:   43252300611,
	}
	f.variantDealActive = &data.ProductVariant{
		ProductID: f.productDealActive.ID,
		PlaceID:   f.testStore.ID,
		Price:     10,
		Limits:    15,
		OfferID:   43252300611,
	}
	f.variantDealExpired = &data.ProductVariant{
		ProductID: f.productDealExpired.ID,
		PlaceID:   f.testStore.ID,
		Price:     10,
		Limits:    15,
		OfferID:   43251994371,
	}

	assert.NoError(t, data.DB.Save(f.variantInStock))
	assert.NoError(t, data.DB.Save(f.variantNotInStock))
	assert.NoError(t, data.DB.Save(f.variantNotInStockRemotely))
	assert.NoError(t, data.DB.Save(f.variantWithDiscount))
	assert.NoError(t, data.DB.Save(f.variantDealActive))
	assert.NoError(t, data.DB.Save(f.variantDealExpired))
}

func (f *fixture) setupDeal(t *testing.T) {
	yesterday := time.Now().AddDate(0, 0, -1)
	tomorrow := time.Now().AddDate(0, 0, 1)

	f.dealActive = &data.Deal{
		Code:       "VALID_DEAL",
		MerchantID: f.testStore.ID,
		Value:      10,
		ExternalID: 306976555063,
		StartAt:    &yesterday,
		EndAt:      &tomorrow,
		Status:     data.DealStatusActive,
	}

	f.dealExpired = &data.Deal{
		Code:       "TEST_EXPIRED_DISCOUNT",
		MerchantID: f.testStore.ID,
		Value:      1,
		ExternalID: 307041402935,
		StartAt:    &yesterday,
		EndAt:      &yesterday,
		Status:     data.DealStatusInactive,
	}
	assert.NoError(t, data.DB.Save(f.dealActive))
	assert.NoError(t, data.DB.Save(f.dealExpired))
}

func (f *fixture) linkProductsWithDeal(t *testing.T) {
	dpValid := data.DealProduct{
		DealID:    f.dealActive.ID,
		ProductID: f.productDealActive.ID,
	}

	dpExpired := data.DealProduct{
		DealID:    f.dealExpired.ID,
		ProductID: f.productDealExpired.ID,
	}

	assert.NoError(t, data.DB.DealProduct.Create(dpValid))
	assert.NoError(t, data.DB.DealProduct.Create(dpExpired))
}

func (f *fixture) SetupData(t *testing.T, apiURL string) {
	f.apiURL = apiURL

	f.validAddress = &data.CartAddress{
		FirstName:    "Test",
		LastName:     "User",
		Address:      "180 John Street",
		AddressOpt:   "",
		City:         "Toronto",
		Country:      "Canada",
		CountryCode:  "CA",
		Province:     "Ontario",
		ProvinceCode: "ON",
		Zip:          "M2J3J3",
	}

	f.setupUser(t)
	f.setupTestStores(t)
	f.setupProduct(t)
	f.setupDeal(t)
	f.linkProductsWithDeal(t)
}

func (f *fixture) TeardownData(t *testing.T) {
	data.DB.Exec("TRUNCATE users cascade;")
	data.DB.Exec("TRUNCATE places cascade;")
}

type MockFacebook struct{}

func (f *MockFacebook) Login(token, inviteCode string) (*data.User, error) {
	username := fmt.Sprintf("test%s@localyyz.com", token[len(token)-1:])
	user := data.User{ID: 0, Network: "facebook", Email: username, Username: username, Name: username}
	return &user, nil
}

func (f *MockFacebook) GetUser(u *data.User) error {
	return nil
}
