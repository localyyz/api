package endtoend

import (
	"context"
	"net/http"
	"testing"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/tests"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ExpressTestSuite struct {
	suite.Suite
	*fixture

	env *tests.Env
}

func (suite *ExpressTestSuite) SetupSuite() {
	suite.env = tests.SetupEnv(suite.T())
	suite.fixture = &fixture{}
	suite.SetupData(suite.T(), suite.env.URL)
}

func (suite *ExpressTestSuite) TearDownSuite() {
	suite.TeardownData(suite.T())
}

func (suite *ExpressTestSuite) TearDownTest() {
	data.DB.Exec("TRUNCATE carts cascade;")
}

func (suite *ExpressTestSuite) TestExpressSuccess() {
	user := suite.user
	client := user.client

	{ // verify default cart exists
		ctx := context.Background()
		cart, _, err := client.ExpressCart.Get(ctx)
		suite.NotNil(cart)
		suite.NoError(err)

		_, resp, err := client.ExpressCart.AddItem(ctx, suite.variantInStock)
		suite.NoError(err)
		require.NotNil(suite.T(), resp)
		require.Equal(suite.T(), http.StatusOK, resp.StatusCode)

		_, _, err = client.ExpressCart.UpdateShippingAddress(
			ctx,
			&data.CartAddress{
				Address:   "123 Toronto Street",
				FirstName: "Someone",
				LastName:  "Localyyz",
				City:      "Toronto",
				Country:   "Canada",
				Province:  "Ontario",
				Zip:       "M5J 1B7",
			})
		suite.NoError(err)

		_, _, err = client.ExpressCart.UpdateShippingMethod(ctx, "canada_post-DOM.EP-10.47")
		suite.NoError(err)

		cart, _, err = client.ExpressCart.GetShippingRates(ctx)
		suite.NoError(err)
		suite.NotEmpty(cart.ShippingRates)
	}

	{ // pay.
		ctx := context.Background()
		cart, _, err := client.ExpressCart.Pay(
			ctx,
			&data.CartAddress{
				FirstName:    "Test",
				LastName:     "Test",
				Address:      "12 Deerford Road",
				AddressOpt:   "",
				City:         "Toronto",
				Country:      "Canada",
				CountryCode:  "CA",
				Province:     "Ontario",
				ProvinceCode: "ON",
				Zip:          "M2J3J3",
			},
			"tok_ca",
			"waseef@localyyz.com",
		)
		require.NoError(suite.T(), err)

		// validate cart
		suite.NotNil(cart)
		suite.Equal(data.CartStatusPaymentSuccess, cart.Status)
	}
}

func TestExpressTestSuite(t *testing.T) {
	suite.Run(t, new(ExpressTestSuite))
}
