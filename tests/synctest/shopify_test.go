package synctest

import (
	"context"
	"strconv"
	"testing"
	"time"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/lib/sync"
	"bitbucket.org/moodie-app/moodie-api/tests"
	"github.com/pressly/lg"
	"github.com/stretchr/testify/suite"
	db "upper.io/db.v3"
)

type SyncTestSuite struct {
	suite.Suite
	*fixture

	env *tests.Env
}

func (suite *SyncTestSuite) SetupSuite() {
	suite.env = tests.SetupEnv(suite.T())
	suite.fixture = &fixture{}
	suite.SetupData(suite.T())
}

func (suite *SyncTestSuite) TearDownSuite() {
	suite.TeardownData(suite.T())
}

func (suite *SyncTestSuite) SetupTest() {
	ctx := context.WithValue(context.Background(), "sync.place", suite.testStore)
	ctx = context.WithValue(ctx, "sync.list", suite.createListing)
	listener := make(sync.Listener)
	ctx = context.WithValue(ctx, sync.SyncListenerCtxKey, listener)

	start := time.Now()
	if err := sync.ShopifyProductListingsCreate(ctx); err != nil {
		if err != sync.ErrProductExist {
			suite.NoError(err)
		}
	} else {
		<-listener
	}
	lg.Debugf("setup test timeit: %s", time.Since(start))
}

func (suite *SyncTestSuite) TearDownTest() {
	data.DB.Exec("TRUNCATE products cascade;")
}

func (suite *SyncTestSuite) TestSyncProductListingCreate() {
	// now check for what's inserted.
	product, err := data.DB.Product.FindOne(
		db.Cond{
			"place_id":    suite.testStore.ID,
			"external_id": suite.createListing[0].ProductID,
		},
	)
	suite.NoError(err)
	suite.NotNil(product)

	// check product syncs
	suite.Equal("Balan Pant in Linen", product.Title)
	suite.Equal(313.60, product.Price)
	suite.Equal(data.ProductStatusApproved, product.Status)
	suite.Equal(data.ProductGenderFemale, product.Gender, "gender")
	suite.Equal(data.CategoryApparel, product.Category.Type, "category type")
	suite.Equal("pants", product.Category.Value, "category value")
	suite.EqualValues(4, product.Score, "score")

	{ // check variant syncs
		variants, err := data.DB.ProductVariant.FindByProductID(product.ID)
		suite.NoError(err)
		suite.NotEmpty(variants)

		suite.Equal(len(suite.createListing[0].Variants), len(variants))
		variantMap := make(map[int64]*shopify.ProductVariant)
		for _, v := range suite.createListing[0].Variants {
			variantMap[v.ID] = v
		}

		for _, v := range variants {
			suite.EqualValues(atof(variantMap[v.OfferID].Price), v.Price)
			suite.EqualValues(atof(variantMap[v.OfferID].CompareAtPrice), v.PrevPrice)
			suite.EqualValues(variantMap[v.OfferID].InventoryQuantity, v.Limits)
		}
	}

	{ // check image syncs
		images, err := data.DB.ProductImage.FindByProductID(product.ID)
		suite.NoError(err)
		suite.NotEmpty(images)

		suite.Equal(len(suite.createListing[0].Images), len(images))
		imageMap := make(map[int64]*shopify.ProductImage)
		for _, m := range suite.createListing[0].Images {
			imageMap[m.ID] = m
		}

		for _, m := range images {
			suite.Contains(imageMap[m.ExternalID].Src, m.ImageURL)
			suite.EqualValues(imageMap[m.ExternalID].Width, m.Width)
			suite.EqualValues(imageMap[m.ExternalID].Height, m.Height)
			suite.EqualValues(imageMap[m.ExternalID].Position, m.Ordering)
			suite.EqualValues(1, m.Score)
		}
	}
}

func (suite *SyncTestSuite) TestSyncProductListingUpdate() {
	ctx := context.WithValue(context.Background(), "sync.place", suite.testStore)
	ctx = context.WithValue(ctx, "sync.list", suite.updateListing)
	listener := make(sync.Listener)
	ctx = context.WithValue(ctx, sync.SyncListenerCtxKey, listener)
	suite.NoError(sync.ShopifyProductListingsUpdate(ctx))
	<-listener

	// now check for what's inserted.
	product, err := data.DB.Product.FindOne(
		db.Cond{
			"place_id":    suite.testStore.ID,
			"external_id": suite.updateListing[0].ProductID,
		},
	)
	suite.NoError(err)
	suite.NotNil(product)

	// check product syncs
	suite.Equal("Balan Pant in Linen", product.Title)
	suite.Equal(313.60, product.Price)
	suite.Equal(data.ProductStatusApproved, product.Status)
	suite.Equal(data.ProductGenderFemale, product.Gender, "gender")
	suite.Equal(data.CategoryApparel, product.Category.Type, "category type")
	suite.Equal("pants", product.Category.Value, "category value")
	suite.EqualValues(4, product.Score, "score")

	{ // check variant syncs
		variants, err := data.DB.ProductVariant.FindByProductID(product.ID)
		suite.NoError(err)
		suite.NotEmpty(variants)

		suite.Equal(len(suite.updateListing[0].Variants), len(variants))
		variantMap := make(map[int64]*shopify.ProductVariant)
		for _, v := range suite.updateListing[0].Variants {
			variantMap[v.ID] = v
		}

		for _, v := range variants {
			suite.EqualValues(atof(variantMap[v.OfferID].Price), v.Price)
			suite.EqualValues(atof(variantMap[v.OfferID].CompareAtPrice), v.PrevPrice)
			suite.EqualValues(variantMap[v.OfferID].InventoryQuantity, v.Limits)
		}
	}

	{ // check image syncs
		images, err := data.DB.ProductImage.FindByProductID(product.ID)
		suite.NoError(err)
		suite.NotEmpty(images)

		suite.Equal(len(suite.updateListing[0].Images), len(images))
		imageMap := make(map[int64]*shopify.ProductImage)
		for _, m := range suite.updateListing[0].Images {
			imageMap[m.ID] = m
		}

		for _, m := range images {
			suite.Contains(imageMap[m.ExternalID].Src, m.ImageURL)
			suite.EqualValues(imageMap[m.ExternalID].Width, m.Width)
			suite.EqualValues(imageMap[m.ExternalID].Height, m.Height)
			suite.EqualValues(imageMap[m.ExternalID].Position, m.Ordering)
			suite.EqualValues(1, m.Score)
		}
	}
}

func (suite *SyncTestSuite) TestSyncProductListingSwapImage() {
	ctx := context.WithValue(context.Background(), "sync.place", suite.testStore)
	ctx = context.WithValue(ctx, "sync.list", suite.swapImageUpdate)
	listener := make(sync.Listener)
	ctx = context.WithValue(ctx, sync.SyncListenerCtxKey, listener)
	suite.NoError(sync.ShopifyProductListingsUpdate(ctx))
	<-listener

	// now check for what's inserted.
	product, err := data.DB.Product.FindOne(
		db.Cond{
			"place_id":    suite.testStore.ID,
			"external_id": suite.swapImageUpdate[0].ProductID,
		},
	)
	suite.NoError(err)
	suite.NotNil(product)

	{ // check image syncs
		images, err := data.DB.ProductImage.FindByProductID(product.ID)
		suite.NoError(err)
		suite.NotEmpty(images)

		suite.Equal(len(suite.swapImageUpdate[0].Images), len(images))
		imageMap := make(map[int64]*shopify.ProductImage)
		for _, m := range suite.swapImageUpdate[0].Images {
			imageMap[m.ID] = m
		}
		for _, m := range images {
			suite.Contains(imageMap[m.ExternalID].Src, m.ImageURL)
			suite.EqualValues(imageMap[m.ExternalID].Width, m.Width)
			suite.EqualValues(imageMap[m.ExternalID].Height, m.Height)
			suite.EqualValues(imageMap[m.ExternalID].Position, m.Ordering)
			suite.EqualValues(1, m.Score)
		}
	}
}

func (suite *SyncTestSuite) TestSyncProductListingOutofStock() {
	ctx := context.WithValue(context.Background(), "sync.place", suite.testStore)
	ctx = context.WithValue(ctx, "sync.list", suite.outOfStock)
	listener := make(sync.Listener)
	ctx = context.WithValue(ctx, sync.SyncListenerCtxKey, listener)
	suite.NoError(sync.ShopifyProductListingsUpdate(ctx))
	<-listener

	// now check for what's inserted.
	product, err := data.DB.Product.FindOne(
		db.Cond{
			"place_id":    suite.testStore.ID,
			"external_id": suite.updateListing[0].ProductID,
		},
	)
	suite.NoError(err)
	suite.NotNil(product)

	// check product syncs
	suite.Equal(data.ProductStatusOutofStock, product.Status)
}

func (suite *SyncTestSuite) TestSyncProductListingRestock() {
	ctx := context.WithValue(context.Background(), "sync.place", suite.testStore)
	{
		ctx = context.WithValue(ctx, "sync.list", suite.outOfStock)
		listener := make(sync.Listener)
		ctx = context.WithValue(ctx, sync.SyncListenerCtxKey, listener)
		suite.NoError(sync.ShopifyProductListingsUpdate(ctx))
		<-listener
	}

	{
		ctx = context.WithValue(ctx, "sync.list", suite.updateListing)
		listener := make(sync.Listener)
		ctx = context.WithValue(ctx, sync.SyncListenerCtxKey, listener)
		suite.NoError(sync.ShopifyProductListingsUpdate(ctx))
		<-listener
	}

	// now check for what's inserted.
	product, err := data.DB.Product.FindOne(
		db.Cond{
			"place_id":    suite.testStore.ID,
			"external_id": suite.updateListing[0].ProductID,
		},
	)
	suite.NoError(err)
	suite.NotNil(product)

	// check product syncs
	suite.Equal(data.ProductStatusApproved, product.Status)
}

func atof(a string) float64 {
	f, _ := strconv.ParseFloat(a, 64)
	return f
}

func TestSyncTestSuite(t *testing.T) {
	suite.Run(t, new(SyncTestSuite))
}
