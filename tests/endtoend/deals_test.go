package endtoend

import (
	"context"
	"net/http"
	"testing"
	"time"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
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
	data.DB.Exec("DELETE FROM deals WHERE user_id IS NOT NULL;")
}

// E2E deals tests

func (suite *DealsTestSuite) TestExpired() {
	user := suite.user
	client := user.client

	_, _, err := client.ExpressCart.AddItem(context.Background(), suite.variantDealExpired)
	suite.NoError(err, "unexpected error")

	cart, _, err := client.ExpressCart.Get(context.Background())
	suite.NoError(err, "unexpected error")

	suite.EqualValues(0, cart.TotalDiscount)
}

func (suite *DealsTestSuite) TestActivateActive() {
	user := suite.user
	client := user.client

	payload := &deals.ActivateRequest{
		DealID:   suite.dealActive.ID,
		StartAt:  data.GetTimeUTCPointer(),
		Duration: 1,
	}
	// Should error. cannot activate a current active deal
	_, resp, err := client.Deal.Activate(context.Background(), payload)
	suite.Equal(http.StatusBadRequest, resp.StatusCode)
	require.Error(suite.T(), err)
}

func (suite *DealsTestSuite) TestActivate() {
	user := suite.user
	client := user.client

	duration := int64(2)
	startAt := data.GetTimeUTCPointer()
	endAt := startAt.Add(time.Duration(duration) * time.Hour)
	parentDeal := suite.dealExpired

	{
		payload := &deals.ActivateRequest{
			DealID:   parentDeal.ID,
			StartAt:  startAt,
			Duration: duration,
		}
		deal, resp, err := client.Deal.Activate(context.Background(), payload)
		suite.Equal(http.StatusCreated, resp.StatusCode)
		require.NoError(suite.T(), err)
		require.NotEmpty(suite.T(), deal)

		suite.Equal(data.DealStatusActive, deal.Status, "unexpected status")
		suite.WithinDuration(startAt.UTC(), deal.StartAt.UTC(), time.Second, "unexpected start time")
		suite.WithinDuration(endAt.UTC(), deal.EndAt.UTC(), time.Second, "unexpected end time")
	}

	var activeDeal *presenter.Deal
	{ // fetch active, deal should include expired but user active deal
		ctx := context.Background()
		deals, resp, err := client.Deal.ListActive(ctx)
		suite.Equal(http.StatusOK, resp.StatusCode)
		require.NoError(suite.T(), err)
		require.NotEmpty(suite.T(), deals)

		for _, d := range deals {
			if d.ParentID != nil &&
				d.UserID != nil &&
				*d.ParentID == parentDeal.ID &&
				*d.UserID == user.ID {

				activeDeal = d
				// validate deal has the right expiry
				suite.WithinDuration(startAt.UTC(), d.StartAt.UTC(), time.Second)
				suite.WithinDuration(endAt.UTC(), d.EndAt.UTC(), time.Second)
			}
		}
		require.NotNil(suite.T(), activeDeal)
		require.NotEmpty(suite.T(), activeDeal.Products)
		require.NotEmpty(suite.T(), activeDeal.Products[0].Variants)
	}

	{ //test purchasing expired deal
		ctx := context.Background()
		var (
			cart *presenter.Cart
			err  error
		)

		dealVariant := activeDeal.Products[0].Variants[0].ProductVariant
		_, resp, err := client.ExpressCart.AddItem(ctx, dealVariant)
		suite.NoError(err)
		require.NotNil(suite.T(), resp)
		require.Equal(suite.T(), http.StatusOK, resp.StatusCode)

		_, _, err = client.ExpressCart.UpdateShippingAddress(ctx, suite.validAddress)
		require.NoError(suite.T(), err)

		cart, _, err = client.ExpressCart.GetShippingRates(ctx)
		require.NoError(suite.T(), err)
		require.NotEmpty(suite.T(), cart.ShippingRates)

		_, _, err = client.ExpressCart.UpdateShippingMethod(ctx, cart.ShippingRates[0].Handle)
		require.NoError(suite.T(), err)

		cart, _, err = client.ExpressCart.Pay(
			ctx,
			suite.validAddress,
			"tok_ca", //test token
			user.Email,
		)
		require.NoError(suite.T(), err)
		require.NotNil(suite.T(), cart)
		require.Equal(suite.T(), data.CartStatusPaymentSuccess, cart.Status)

		// validate that the discount was applied
		suite.NotZero(cart.TotalDiscount, "unexpected discount zero")
		suite.EqualValues(parentDeal.Value*100.0, cart.TotalDiscount, "unexpected discount value")
	}
}

// TODO:
//func (suite *DealsTestSuite) TestCapHit() {
//// start with success
//suite.TestSuccess()

//user2 := suite.user2
//client := user2.client

//_, resp, err := client.ExpressCart.AddItem(context.Background(), suite.variantDealValid)
//require.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode)
//require.Contains(suite.T(), err.Error(), "the products from this lightning collection have been sold out")
//}

// TODO:
//func (suite *DealsTestSuite) TestUserLimit() {
//// start with success
//suite.TestSuccess()

//user := suite.user
//client := user.client

//_, resp, err := client.ExpressCart.AddItem(context.Background(), suite.variantDealValid)
//require.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode)
//require.Contains(suite.T(), err.Error(), "you have already purchased today's deal")
//}

func TestDealsTestSuite(t *testing.T) {
	suite.Run(t, new(DealsTestSuite))
}
