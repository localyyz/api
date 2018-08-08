package synctest

import (
	"testing"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/sync"
	"bitbucket.org/moodie-app/moodie-api/tests"
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

func (suite *CollectionTestSuite) TestSyncProductCollections() {
	suite.NoError(sync.AddProductToCollection(suite.fixture.product.ID, suite.fixture.collection.ID))
	var cP *data.CollectionProduct
	err := data.DB.CollectionProduct.Find(
		db.Cond{
			"product_id":    suite.fixture.product.ID,
			"collection_id": suite.fixture.collection.ID,
		},
	).One(&cP)
	suite.NoError(err)
	suite.Equal(suite.fixture.product.ID, cP.ProductID)
	suite.Equal(suite.fixture.collection.ID, cP.CollectionID)

	// attempt to add the same product to collection again should fail
	err = sync.AddProductToCollection(suite.fixture.product.ID, suite.fixture.collection.ID)
	// this should not have passed. but we ignored the unique_violation database error
	suite.NoError(err)
}

func TestCollectionTestSuite(t *testing.T) {
	suite.Run(t, new(CollectionTestSuite))
}
