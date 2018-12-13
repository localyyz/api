package synctest

import (
	"context"
	"testing"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/sync"
	"bitbucket.org/moodie-app/moodie-api/tests"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	db "upper.io/db.v3"
)

type SyncParseTestSuite struct {
	suite.Suite
	*fixture

	env *tests.Env
}

func (suite *SyncParseTestSuite) SetupSuite() {
	suite.env = tests.SetupEnv(suite.T())
	suite.fixture = &fixture{}
	suite.SetupData(suite.T())
	require.NoError(suite.T(), sync.SetupCache())
}

func (suite *SyncParseTestSuite) TearDownSuite() {
	suite.TeardownData(suite.T())
}

func (suite *SyncParseTestSuite) TearDownTest() {
	data.DB.Exec("TRUNCATE products cascade;")
}

// test gender parsing
func (suite *SyncParseTestSuite) TestSyncProductListingParseGender() {
	ctx := context.WithValue(context.Background(), "sync.place", suite.testStoreFemale)
	{
		ctx = context.WithValue(ctx, "sync.list", suite.parse1)
		listener := make(sync.Listener)
		ctx = context.WithValue(ctx, sync.SyncListenerCtxKey, listener)
		suite.NoError(sync.ShopifyProductListingsCreate(ctx))
		<-listener
	}

	// now check for what's inserted.
	product, err := data.DB.Product.FindOne(
		db.Cond{
			"place_id":    suite.testStoreFemale.ID,
			"external_id": suite.parse1[0].ProductID,
		},
	)
	suite.NoError(err)
	suite.NotNil(product)

	// validate product is correctly labeled as male.
	suite.Equal(data.ProductGenderMale, product.Gender)
}

func TestSyncParseTestSuite(t *testing.T) {
	suite.Run(t, new(SyncParseTestSuite))
}
