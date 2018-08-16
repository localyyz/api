package endtoend

import (
	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/tests"
	"context"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

type UserCollTestSuite struct {
	suite.Suite
	*fixture

	env *tests.Env
}

func (suite *UserCollTestSuite) SetupSuite() {
	suite.env = tests.SetupEnv(suite.T())

	suite.TeardownData(suite.T())
	suite.fixture = &fixture{}
}

func (suite *UserCollTestSuite) TearDownSuite() {
	suite.TeardownData(suite.T())
}

func (suite *UserCollTestSuite) TearDownTest() {
	data.DB.Exec("TRUNCATE users cascade;")
}

func (suite *UserCollTestSuite) SetupTest() {
	suite.SetupData(suite.T(), suite.env.URL)
}

// test if a new collection is created
func (suite *UserCollTestSuite) TestCreateNewColl() {
	user := suite.user
	client := user.client
	ctx := context.Background()

	// create the collection
	collection, resp, err := client.UserColl.CreateNewUserColl(ctx, "Summer 2018")
	suite.NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)
	suite.Equal("Summer 2018", collection.Title)
}

// test if a collection is successfully deleted
func (suite *UserCollTestSuite) TestDeleteColl() {
	user := suite.user
	client := user.client
	ctx := context.Background()

	// create the collection
	collection, resp, err := client.UserColl.CreateNewUserColl(ctx, "Summer 2018")
	suite.NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)
	suite.Equal("Summer 2018", collection.Title)

	// delete the collection
	resp, err = client.UserColl.DeleteUserColl(ctx, collection.ID)
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)

	// make sure it returns nothing
	collections, resp, err := client.UserColl.ListUserColl(ctx)
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
	suite.Equal(0, len(collections))
}

// test if a collection is updated
func (suite *UserCollTestSuite) TestUpdateColl() {
	user := suite.user
	client := user.client
	ctx := context.Background()

	// create the collection
	collection, resp, err := client.UserColl.CreateNewUserColl(ctx, "Summer 2018")
	suite.NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)
	suite.Equal("Summer 2018", collection.Title)

	// update the collection
	collection, resp, err = client.UserColl.UpdateUserColl(ctx, collection.ID, "Summer 2019")
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
	suite.Equal("Summer 2019", collection.Title)

	// get the collections
	collections, resp, err := client.UserColl.ListUserColl(ctx)
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
	suite.Equal("Summer 2019", collections[0].Title)
}

// test if it successfully returns a users single collection
func (suite *UserCollTestSuite) TestGetUserColl() {
	user := suite.user
	client := user.client
	ctx := context.Background()

	// create the collection
	collection, resp, err := client.UserColl.CreateNewUserColl(ctx, "Summer 2018")
	suite.NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)
	suite.Equal("Summer 2018", collection.Title)

	// get the individual collection
	collection, resp, err = client.UserColl.GetUserColl(ctx, collection.ID)
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
	suite.Equal("Summer 2018", collection.Title)

	// make sure there are no products added
	products, resp, err := client.UserColl.ListUserCollProd(ctx, collection.ID)
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
	suite.Equal(0, len(products))
}

// test if it successfully lists all the users collections
func (suite *UserCollTestSuite) TestListUserColl() {
	user := suite.user
	client := user.client
	ctx := context.Background()

	// create the multiple collections
	collection, resp, err := client.UserColl.CreateNewUserColl(ctx, "Summer 2018")
	suite.NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)
	suite.Equal("Summer 2018", collection.Title)

	collection, resp, err = client.UserColl.CreateNewUserColl(ctx, "Summer 2019")
	suite.NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)
	suite.Equal("Summer 2019", collection.Title)

	collection, resp, err = client.UserColl.CreateNewUserColl(ctx, "Summer 2020")
	suite.NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)
	suite.Equal("Summer 2020", collection.Title)

	// read the collections
	collections, resp, err := client.UserColl.ListUserColl(ctx)
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
	suite.Equal(3, len(collections))
	suite.Equal("Summer 2020", collections[0].Title)
	suite.Equal("Summer 2019", collections[1].Title)
	suite.Equal("Summer 2018", collections[2].Title)

}

func (suite *UserCollTestSuite) TestAddProductToColl() {
	user := suite.user
	client := user.client
	ctx := context.Background()

	// create the collection
	collection, resp, err := client.UserColl.CreateNewUserColl(ctx, "Summer 2018")
	suite.NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)
	suite.Equal("Summer 2018", collection.Title)

	// create the product in the collection
	product, resp, err := client.UserColl.CreateProdInUserColl(ctx, collection.ID, suite.productInStock.ID)
	suite.NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)
	suite.Equal(suite.productInStock.ID, product.ID)

	// list the products
	products, resp, err := client.UserColl.ListUserCollProd(ctx, collection.ID)
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
	suite.Equal(1, len(products))
	suite.Equal(suite.productInStock.ID, products[0].ID)

}

func (suite *UserCollTestSuite) TestRemoveProdFromColl() {
	user := suite.user
	client := user.client
	ctx := context.Background()

	// create the collection
	collection, resp, err := client.UserColl.CreateNewUserColl(ctx, "Summer 2018")
	suite.NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)
	suite.Equal("Summer 2018", collection.Title)

	// create the product
	product, resp, err := client.UserColl.CreateProdInUserColl(ctx, collection.ID, suite.productInStock.ID)
	suite.NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)
	suite.Equal(suite.productInStock.ID, product.ID)

	// list the products
	products, resp, err := client.UserColl.ListUserCollProd(ctx, collection.ID)
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
	suite.Equal(1, len(products))
	suite.Equal(suite.productInStock.ID, products[0].ID)

	// delete from the collection
	resp, err = client.UserColl.DeleteProdFromUserColl(ctx, collection.ID, products[0].ID)
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)

	// list all the products make sure no more products
	products, resp, err = client.UserColl.ListUserCollProd(ctx, collection.ID)
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
	suite.Equal(0, len(products))
}

func (suite *UserCollTestSuite) TestRemoveProdFromAllColl() {
	user := suite.user
	client := user.client
	ctx := context.Background()

	// create the first collection
	collection1, resp, err := client.UserColl.CreateNewUserColl(ctx, "Summer 2018")
	suite.NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)
	suite.Equal("Summer 2018", collection1.Title)

	// add the product to the first collection
	product, resp, err := client.UserColl.CreateProdInUserColl(ctx, collection1.ID, suite.productInStock.ID)
	suite.NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)
	suite.Equal(suite.productInStock.ID, product.ID)

	// create the second collection
	collection2, resp, err := client.UserColl.CreateNewUserColl(ctx, "Summer 2019")
	suite.NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)
	suite.Equal("Summer 2019", collection2.Title)

	// add the same product to the second collection
	product, resp, err = client.UserColl.CreateProdInUserColl(ctx, collection2.ID, suite.productInStock.ID)
	suite.NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)
	suite.Equal(suite.productInStock.ID, product.ID)

	// delete the product from all the collections
	resp, err = client.UserColl.DeleteProdFromAllUserColl(ctx, product.ID)
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)

	// get products for the first collection make sure length is 0
	products, resp, err := client.UserColl.ListUserCollProd(ctx, collection1.ID)
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
	suite.Equal(0, len(products))

	// get products for the second collection make sure length is 0
	products, resp, err = client.UserColl.ListUserCollProd(ctx, collection2.ID)
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
	suite.Equal(0, len(products))
}

// make sure one user cant access another user's collections
func (suite *UserCollTestSuite) TestPrivateColl() {
	user := suite.user
	client := user.client
	ctx := context.Background()

	collection, resp, err := client.UserColl.CreateNewUserColl(ctx, "Summer 2018")
	suite.NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)
	suite.Equal("Summer 2018", collection.Title)

	user2 := suite.user2
	client2 := user2.client

	collection, resp, err = client2.UserColl.GetUserColl(ctx, collection.ID)
	suite.Equal(http.StatusBadRequest, resp.StatusCode)
}

// make sure deleting from one user does not affect other users
func (suite *UserCollTestSuite) TestDeleteDoesNotAffectOtherUser() {
	user := suite.user
	client := user.client
	ctx := context.Background()

	// create the first users collection
	collection1, resp, err := client.UserColl.CreateNewUserColl(ctx, "Summer 2018")
	suite.NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)
	suite.Equal("Summer 2018", collection1.Title)

	// add the product to the first users collection
	product, resp, err := client.UserColl.CreateProdInUserColl(ctx, collection1.ID, suite.productInStock.ID)
	suite.NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)
	suite.Equal(suite.productInStock.ID, product.ID)

	user2 := suite.user2
	client2 := user2.client

	// create the second users collection
	collection2, resp, err := client2.UserColl.CreateNewUserColl(ctx, "Summer 2018")
	suite.NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)
	suite.Equal("Summer 2018", collection2.Title)

	// add the same product to the second users collection
	product, resp, err = client2.UserColl.CreateProdInUserColl(ctx, collection2.ID, suite.productInStock.ID)
	suite.NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)
	suite.Equal(suite.productInStock.ID, product.ID)

	// delete the product from all colelctions for the first user
	resp, err = client.UserColl.DeleteProdFromAllUserColl(ctx, suite.productInStock.ID)
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)

	// get the products for the first user - make sure length is 0
	products, resp, err := client.UserColl.ListUserCollProd(ctx, collection1.ID)
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
	suite.Equal(0, len(products))

	// get the products for the second user - make sure length is 1
	products, resp, err = client2.UserColl.ListUserCollProd(ctx, collection2.ID)
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
	suite.Equal(1, len(products))
	suite.Equal(suite.productInStock.ID, products[0].ID)
}
func TestUserCollSuite(t *testing.T) {
	suite.Run(t, new(UserCollTestSuite))
}
