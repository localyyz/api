package endtoend

import (
	"context"
	"net/http"
	"testing"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopper"
	"bitbucket.org/moodie-app/moodie-api/tests"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	db "upper.io/db.v3"
)

type CheckoutAnonTestSuite struct {
	suite.Suite
	*fixture

	env *tests.Env
}

func (suite *CheckoutAnonTestSuite) SetupSuite() {
	suite.env = tests.SetupEnv(suite.T())
	suite.fixture = &fixture{}
	suite.SetupData(suite.T(), suite.env.URL)
}

func (suite *CheckoutAnonTestSuite) TearDownSuite() {
	suite.TeardownData(suite.T())
}

func (suite *CheckoutAnonTestSuite) TearDownTest() {
	data.DB.Exec("TRUNCATE carts cascade;")
}

// E2E checkout tests

// anonUser should be able to add to cart successfully.
func (suite *CheckoutAnonTestSuite) TestAnonAddtoCart() {
	user := suite.anonUser
	client := user.client
	ctx := context.Background()

	_, resp, err := client.Cart.AddItem(ctx, &data.CartItem{VariantID: suite.variantInStock.ID})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), http.StatusCreated, resp.StatusCode)
}

// test: anonUser checkout success and transition into full user
func (suite *CheckoutAnonTestSuite) TestSuccessTransition() {
	user := suite.anonUser
	client := user.client
	ctx := context.Background()

	cartEmail := "testanon@localyyz.com"

	// add item to the cart
	_, resp, err := client.Cart.AddItem(ctx, &data.CartItem{VariantID: suite.variantInStock.ID})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), http.StatusCreated, resp.StatusCode)

	// update the cart with addresses and email
	_, _, err = client.Cart.Put(ctx, &data.Cart{
		ShippingAddress: suite.validAddress,
		BillingAddress:  suite.validAddress,
		Email:           cartEmail,
	})
	require.NoError(suite.T(), err)

	// checkout
	_, _, err = client.Cart.Checkout(ctx)
	require.NoError(suite.T(), err)

	// pay
	paymentCard := &shopper.PaymentCard{
		Number: "4242424242424242",
		Expiry: "12/28",
		Name:   "Someone Localyyz",
		CVC:    "123",
	}
	cart, _, err := client.Cart.Pay(ctx, paymentCard)
	require.NoError(suite.T(), err)
	suite.Equal(data.CartStatusComplete, cart.Status)

	// verify that user is still anonymous
	u, err := data.DB.User.FindByID(user.ID)
	require.NoError(suite.T(), err)
	suite.Equal("shadow", u.Network)
	suite.Equal(cartEmail, u.Email)

	// attempt to login with user email and some password should return
	// unauthorized
	pswd := "localyyz best ever"
	_, resp, err = client.User.LoginWithEmail(ctx, cartEmail, pswd)
	require.Error(suite.T(), err)
	suite.Equal(http.StatusUnauthorized, resp.StatusCode)

	// signup
	_, resp, err = client.User.SignupWithEmail(ctx, cartEmail, "localyyz test", pswd)
	require.NoError(suite.T(), err)
	suite.Equal(http.StatusCreated, resp.StatusCode)

	// verify the user is now a new email user
	// verify that user is still anonymous
	u, err = data.DB.User.FindByID(user.ID)
	require.NoError(suite.T(), err)
	suite.Equal("email", u.Network)
	suite.Equal(cartEmail, u.Email)
	suite.Equal(cartEmail, u.Username)
	suite.NotEmpty(u.PasswordHash)

	// verify that the cart is still attached to the anon user
	_, err = data.DB.Cart.FindOne(db.Cond{"user_id": u.ID, "id": cart.ID})
	require.NoError(suite.T(), err)
}

func TestCheckoutAnonTestSuite(t *testing.T) {
	suite.Run(t, new(CheckoutAnonTestSuite))
}
