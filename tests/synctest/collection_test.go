package synctest

import (
	"context"
	"testing"
	"time"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/sync"
	"bitbucket.org/moodie-app/moodie-api/tests"
	"github.com/pressly/lg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"upper.io/db.v3"
)

type CollectionTestSuite struct {
	suite.Suite
	*fixture

	env *tests.Env
}

func (suite *CollectionTestSuite) SetupSuite() {
	suite.env = tests.SetupEnv(suite.T())
	suite.fixture = &fixture{}
	suite.SetupData(suite.T())
}

func (suite *CollectionTestSuite) TearDownSuite() {
	suite.TeardownData(suite.T())
}

func (suite *CollectionTestSuite) SetupTest() {
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

func (suite *CollectionTestSuite) TearDownTest() {

}

func (suite *CollectionTestSuite) TestSyncProductCollectionsError() {
	suite.NoError(sync.AddProductToCollection(suite.fixture.ProductNotUnique.ID, suite.fixture.CollectionNotUnique.ID))
	var cP *data.CollectionProduct
	err := data.DB.CollectionProduct.Find(db.Cond{"product_id": suite.fixture.ProductNotUnique.ID, "collection_id": suite.fixture.CollectionNotUnique.ID}).One(&cP)
	if err != nil && err != db.ErrNoMoreRows {
		assert.Fail(suite.T(), "Added product to collection violation unique constraint")
	}
}

func (suite *CollectionTestSuite) TestSyncProductCollections() {
	suite.NoError(sync.AddProductToCollection(suite.fixture.ProductUnique.ID, suite.fixture.CollectionUnique.ID))
	var cP *data.CollectionProduct
	err := data.DB.CollectionProduct.Find(db.Cond{"product_id": suite.fixture.ProductUnique.ID, "collection_id": suite.fixture.CollectionUnique.ID}).One(&cP)
	suite.NoError(err)
	suite.Equal(cP.ProductID, suite.fixture.ProductUnique.ID)
	suite.Equal(cP.CollectionID, suite.fixture.CollectionUnique.ID)
}

func TestCollectionTestSuite(t *testing.T) {
	suite.Run(t, new(CollectionTestSuite))
}
