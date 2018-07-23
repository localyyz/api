package endtoend

import (
	"context"
	"net/http"
	"testing"
	"time"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/tests"
	"bitbucket.org/moodie-app/moodie-api/web/deals"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type DealsTestSuite struct {
	suite.Suite
	*fixture

	env *tests.Env
}

func (suite *DealsTestSuite) SetupSuite() {
	suite.env = tests.SetupEnv(suite.T())
	suite.fixture = &fixture{}
	suite.SetupData(suite.T(), suite.env.URL)
}

func (suite *DealsTestSuite) TearDownSuite() {
	suite.TeardownData(suite.T())
}

func (suite *DealsTestSuite) TearDownTest() {
	data.DB.Exec("TRUNCATE carts cascade;")
	data.DB.Exec("TRUNCATE user_deals cascade;")
}

// E2E deals tests

func (suite *DealsTestSuite) TestSuccess() {
	user := suite.user
	client := user.client

	{ //add item to cart
		ctx := context.Background()
		_, resp, err := client.ExpressCart.AddItem(ctx, suite.variantDealValid)
		suite.NoError(err)
		require.NotNil(suite.T(), resp)
		require.Equal(suite.T(), http.StatusOK, resp.StatusCode)

		_, _, err = client.ExpressCart.UpdateShippingAddress(
			ctx,
			&data.CartAddress{
				FirstName:    "Test",
				LastName:     "Test",
				Address:      "180 John Street",
				AddressOpt:   "",
				City:         "Toronto",
				Country:      "Canada",
				CountryCode:  "CA",
				Province:     "Ontario",
				ProvinceCode: "ON",
				Zip:          "M2J3J3",
			})
		require.NoError(suite.T(), err)

		_, _, err = client.ExpressCart.UpdateShippingMethod(ctx, "canada_post-DOM.EP-10.47")
		require.NoError(suite.T(), err)
	}
	{ //pay
		ctx := context.Background()
		cart, _, err := client.ExpressCart.Pay(
			ctx,
			&data.CartAddress{
				FirstName:    "Test",
				LastName:     "Test",
				Address:      "180 John Street",
				AddressOpt:   "",
				City:         "Toronto",
				Country:      "Canada",
				CountryCode:  "CA",
				Province:     "Ontario",
				ProvinceCode: "ON",
				Zip:          "M2J3J3",
			},
			"tok_ca", //test token
			"waseef@localyyz.com",
		)
		require.NoError(suite.T(), err)
		require.NotNil(suite.T(), cart)
		require.Equal(suite.T(), data.CartStatusPaymentSuccess, cart.Status)
	}
}

func (suite *DealsTestSuite) TestExpired() {
	user := suite.user
	client := user.client

	_, _, err := client.ExpressCart.AddItem(context.Background(), suite.variantDealExpired)
	suite.Contains(err.Error(), "this lightning deal has ended")
}

func (suite *DealsTestSuite) TestCapHit() {
	// start with success
	suite.TestSuccess()

	user2 := suite.user2
	client := user2.client

	_, resp, err := client.ExpressCart.AddItem(context.Background(), suite.variantDealValid)
	require.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode)
	require.Contains(suite.T(), err.Error(), "the products from this lightning collection have been sold out")
}

func (suite *DealsTestSuite) TestUserLimit() {
	// start with success
	suite.TestSuccess()

	user := suite.user
	client := user.client

	_, resp, err := client.ExpressCart.AddItem(context.Background(), suite.variantDealValid)
	require.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode)
	require.Contains(suite.T(), err.Error(), "you have already purchased today's deal")
}

func (suite *DealsTestSuite) TestActivate() {
	user := suite.user
	client := user.client

	{
		payload := &deals.ActivateRequest{
			DealID:   suite.dealExpired.ID,
			StartAt:  data.GetTimeUTCPointer(),
			Duration: 2,
		}
		deal, resp, err := client.Deal.Activate(context.Background(), payload)
		suite.Equal(http.StatusCreated, resp.StatusCode)
		require.NoError(suite.T(), err)
		require.NotEmpty(suite.T(), deal)

		suite.Equal(data.CollectionStatusActive, deal.Status, "unexpected status")
		suite.Equal(payload.StartAt, deal.StartAt, "unexpected start time")
		suite.Equal(payload.StartAt.Add(time.Duration(payload.Duration)*time.Hour), *deal.EndAt, "unexpected end time")
	}

	{ // fetch active, deal should include expired but user active deal
		ctx := context.Background()
		deals, resp, err := client.Deal.ListActive(ctx)
		suite.Equal(http.StatusOK, resp.StatusCode)
		require.NoError(suite.T(), err)
		require.NotEmpty(suite.T(), deals)

		isFound := false
		for _, d := range deals {
			if d.ID == suite.dealExpired.ID {
				isFound = true
			}
		}
		require.True(suite.T(), isFound)
	}

	// test purchasing expired deal
	{
		//add item to cart
		ctx := context.Background()
		_, resp, err := client.ExpressCart.AddItem(ctx, suite.variantDealExpired)
		suite.NoError(err)
		require.NotNil(suite.T(), resp)
		require.Equal(suite.T(), http.StatusOK, resp.StatusCode)

		_, _, err = client.ExpressCart.UpdateShippingAddress(
			ctx,
			&data.CartAddress{
				FirstName:    "Test",
				LastName:     "Test",
				Address:      "180 John Street",
				AddressOpt:   "",
				City:         "Toronto",
				Country:      "Canada",
				CountryCode:  "CA",
				Province:     "Ontario",
				ProvinceCode: "ON",
				Zip:          "M2J3J3",
			})
		require.NoError(suite.T(), err)

		_, _, err = client.ExpressCart.UpdateShippingMethod(ctx, "canada_post-DOM.EP-10.47")
		require.NoError(suite.T(), err)

		// pay
		cart, _, err := client.ExpressCart.Pay(
			ctx,
			&data.CartAddress{
				FirstName:    "Test",
				LastName:     "Test",
				Address:      "180 John Street",
				AddressOpt:   "",
				City:         "Toronto",
				Country:      "Canada",
				CountryCode:  "CA",
				Province:     "Ontario",
				ProvinceCode: "ON",
				Zip:          "M2J3J3",
			},
			"tok_ca", //test token
			"waseef@localyyz.com",
		)
		require.NoError(suite.T(), err)
		require.NotNil(suite.T(), cart)
		require.Equal(suite.T(), data.CartStatusPaymentSuccess, cart.Status)
	}
}

func TestDealsTestSuite(t *testing.T) {
	suite.Run(t, new(DealsTestSuite))
}
