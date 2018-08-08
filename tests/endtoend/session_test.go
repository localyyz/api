package endtoend

import (
	"context"
	"testing"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/tests"
	"github.com/stretchr/testify/suite"
	"net/http"
	"upper.io/db.v3"
	"github.com/stretchr/testify/require"
	"fmt"
	"bytes"
	"encoding/json"
	"bitbucket.org/moodie-app/moodie-api/web/auth"
)

type SessionTestSuite struct {
	suite.Suite
	*fixture

	env *tests.Env
}

func (suite *SessionTestSuite) SetupSuite() {
	suite.env = tests.SetupEnv(suite.T())

	suite.TeardownData(suite.T())
	suite.fixture = &fixture{}
	suite.SetupData(suite.T(), suite.env.URL)
}

func (suite *SessionTestSuite) TearDownSuite() {
	suite.TeardownData(suite.T())
}

func (suite *SessionTestSuite) TestSessionPublic() {
	// public routes can be accessed without session
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s", suite.env.URL), nil)
	// verify request returns successfully
	rr, _ := http.DefaultClient.Do(req)
	suite.Equal(http.StatusOK, rr.StatusCode)
}

func (suite *SessionTestSuite) TestShadowUser() {
	//when new user opens the app for the first time the backend creates a "shadow" user
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s", suite.env.URL), nil)
	req.Header.Add("X-DEVICE-ID", "test-device-token")

	//verify it returns ok
	rr, _ := http.DefaultClient.Do(req)
	suite.Equal(http.StatusOK, rr.StatusCode)

	//manualy verify if the backend created a shadow user
	user, _ := data.DB.User.FindByUsername("test-device-token")
	suite.Equal("shadow", user.Network)
}

func (suite *SessionTestSuite) TestSessionSemiAuthRouteWithDeviceID() {
	// anonymous user session with device id
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/carts/express", suite.env.URL), nil)
	req.Header.Add("X-DEVICE-ID", "test-device-token")

	// verify default cart is okay
	rr, _ := http.DefaultClient.Do(req)
	suite.Equal(http.StatusOK, rr.StatusCode)
}

// TODO: change this to authorized
func (suite *SessionTestSuite) TestSessionAuthRouteWithDeviceID() {
	// anonymous user session with device id
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/carts/default", suite.env.URL), nil)
	req.Header.Add("X-DEVICE-ID", "test-device-token")

	// verify default cart is Unauthorized
	rr, _ := http.DefaultClient.Do(req)
	suite.Equal(http.StatusOK, rr.StatusCode)
}

func (suite *SessionTestSuite) TestLoginWithEmail(){
	user := suite.user
	client := suite.user.client
	ctx := context.Background()

	authUser, resp, err := client.User.LoginWithEmail(ctx, user.Email, "test1234")
	suite.NoError(err)
	suite.Equal(resp.StatusCode, http.StatusOK)
	suite.Equal(user.ID, authUser.ID)
}

func (suite *SessionTestSuite) TestSessionEmailSignupWithDeviceID() {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(map[string]string{
		"fullName":        "testuser signup",
		"email":           "test@localyyz.com",
		"password":        "test1234",
		"passwordConfirm": "test1234",
	})

	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/signup", suite.env.URL), buf)
	uID := "test-device-token-email-signup"
	req.Header.Add("X-DEVICE-ID", uID)

	// verify new user is created
	rr, _ := http.DefaultClient.Do(req)
	suite.Equal(http.StatusCreated, rr.StatusCode)

	var authUser *auth.AuthUser
	suite.NoError(json.NewDecoder(rr.Body).Decode(&authUser))

	// validate new user
	suite.Equal("test@localyyz.com", authUser.Username)
	suite.Equal("test@localyyz.com", authUser.Email)
	suite.Equal("testuser signup", authUser.Name)
	suite.Equal("email", authUser.Network)
	suite.NotEmpty(authUser.ID)

	dbUser, err := data.DB.User.FindByID(authUser.ID)
	suite.NoError(err)
	suite.NotNil(dbUser.DeviceToken)
	suite.Equal(uID, *dbUser.DeviceToken)
}

func (suite *SessionTestSuite) TestSessionEmailSignup() {

	client := suite.anonUser.client
	ctx := context.Background()

	userCountBefore, _ := data.DB.User.Find(db.Cond{}).Count()

	authUser, resp, err := client.User.SignupWithEmail(ctx, "waseef@localyyz.com")
	suite.NoError(err)
	suite.Equal(resp.StatusCode, http.StatusCreated)
	suite.Equal("waseef@localyyz.com", authUser.Username)
	suite.Equal("waseef@localyyz.com", authUser.Email)
	suite.Equal("email", authUser.Network)
	suite.NotEmpty(authUser.ID)

	userCountAfter, _ := data.DB.User.Find(db.Cond{}).Count()

	// indicates no new entries were added
	suite.Equal(userCountBefore, userCountAfter)

	// indicates we have created a user with this is
	dbUser, err := data.DB.User.FindByID(authUser.ID)
	suite.NoError(err)
	suite.NotNil(dbUser)

	dbUser, err = data.DB.User.FindByID(suite.anonUser.ID)
	if err != nil && err != db.ErrNoMoreRows {
		suite.Fail("Did not replace user")
	}
}

func (suite *SessionTestSuite) TestSessionFacebookSignup() {
	connect.FacebookLogin = &MockFacebook{}

	client := suite.anonUser2.client
	ctx := context.Background()

	userCountBefore, _ := data.DB.User.Find(db.Cond{}).Count()

	authUser, resp, err := client.User.SignupWithFacebook(ctx, "localyyz-test-token-login-3")
	suite.NoError(err)
	suite.Equal(resp.StatusCode, http.StatusOK)
	suite.Equal("facebook", authUser.Network)
	suite.Equal("test3@localyyz.com", authUser.Email)

	userCountAfter, _ := data.DB.User.Find(db.Cond{}).Count()
	suite.Equal(userCountBefore, userCountAfter)

	dbUser, err := data.DB.User.FindByID(authUser.ID)
	suite.NoError(err)
	suite.NotNil(dbUser)

	dbUser, err = data.DB.User.FindByID(suite.anonUser2.ID)
	if err != nil && err != db.ErrNoMoreRows {
		suite.Fail("Did not replace user")
	}

}

func (suite *SessionTestSuite) TestTransitionPaymentUserEmail() {
	user := suite.anonUser3
	client := user.client
	ctx := context.Background()

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

		// fetch the shipping rate (
		// NOTE shopify will error out with "expired shipping_line" error if we
		// dont fetch shipping rate
		cart, _, err = client.ExpressCart.GetShippingRates(ctx)
		require.NoError(suite.T(), err)
		require.NotEmpty(suite.T(), cart.ShippingRates)

		// update shipping method
		_, _, err = client.ExpressCart.UpdateShippingMethod(ctx, cart.ShippingRates[0].Handle)
		suite.NoError(err)
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

	{
		userCountBefore, _ := data.DB.User.Find(db.Cond{}).Count()

		authUser, resp, err := client.User.SignupWithEmail(ctx, "waseef2@localyyz.com")
		suite.NoError(err)

		suite.Equal(resp.StatusCode, http.StatusCreated)
		suite.Equal("waseef2@localyyz.com", authUser.Username)
		suite.Equal("waseef2@localyyz.com", authUser.Email)
		suite.Equal("email", authUser.Network)
		suite.NotEmpty(authUser.ID)

		userCountAfter, _ := data.DB.User.Find(db.Cond{}).Count()

		suite.Equal(userCountBefore, userCountAfter)

		dbUser, err := data.DB.User.FindByID(authUser.ID)
		suite.NoError(err)
		suite.NotNil(dbUser)

		dbUser, err = data.DB.User.FindByID(suite.anonUser3.ID)
		if err != nil && err != db.ErrNoMoreRows {
			suite.Fail("Did not replace user")
		}
	}
}

func (suite *SessionTestSuite) TestTransitionPaymentUserFacebook() {

	connect.FacebookLogin = &MockFacebook{}

	user := suite.anonUser4
	client := user.client
	ctx := context.Background()

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

		// fetch the shipping rate (
		// NOTE shopify will error out with "expired shipping_line" error if we
		// dont fetch shipping rate
		cart, _, err = client.ExpressCart.GetShippingRates(ctx)
		require.NoError(suite.T(), err)
		require.NotEmpty(suite.T(), cart.ShippingRates)

		// update shipping method
		_, _, err = client.ExpressCart.UpdateShippingMethod(ctx, cart.ShippingRates[0].Handle)
		suite.NoError(err)
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

	{
		userCountBefore, _ := data.DB.User.Find(db.Cond{}).Count()

		authUser, resp, err := client.User.SignupWithFacebook(ctx, "localyyz-test-token-login-4")
		suite.NoError(err)
		suite.Equal(resp.StatusCode, http.StatusOK)
		suite.Equal("facebook", authUser.Network)
		suite.Equal("test4@localyyz.com", authUser.Email)

		userCountAfter, _ := data.DB.User.Find(db.Cond{}).Count()
		suite.Equal(userCountBefore, userCountAfter)

		dbUser, err := data.DB.User.FindByID(authUser.ID)
		suite.NoError(err)
		suite.NotNil(dbUser)

		dbUser, err = data.DB.User.FindByID(suite.anonUser4.ID)
		if err != nil && err != db.ErrNoMoreRows {
			suite.Fail("Did not replace user")
		}
	}
}

func (suite *SessionTestSuite) TestBuyWithEmailSignupWithFacebook(){
	connect.FacebookLogin = &MockFacebook{}

	user := suite.anonUser5
	client := user.client
	ctx := context.Background()

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

		// fetch the shipping rate (
		// NOTE shopify will error out with "expired shipping_line" error if we
		// dont fetch shipping rate
		cart, _, err = client.ExpressCart.GetShippingRates(ctx)
		require.NoError(suite.T(), err)
		require.NotEmpty(suite.T(), cart.ShippingRates)

		// update shipping method
		_, _, err = client.ExpressCart.UpdateShippingMethod(ctx, cart.ShippingRates[0].Handle)
		suite.NoError(err)
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

	{
		userCountBefore, _ := data.DB.User.Find(db.Cond{}).Count()

		authUser, resp, err := client.User.SignupWithFacebook(ctx, "localyyz-test-token-login-5")
		suite.NoError(err)
		suite.Equal(resp.StatusCode, http.StatusOK)
		suite.Equal("facebook", authUser.Network)
		suite.Equal("test5@localyyz.com", authUser.Email)

		userCountAfter, _ := data.DB.User.Find(db.Cond{}).Count()
		suite.Equal(userCountBefore, userCountAfter)

		dbUser, err := data.DB.User.FindByID(authUser.ID)
		suite.NoError(err)
		suite.NotNil(dbUser)

		dbUser, err = data.DB.User.FindByID(suite.anonUser5.ID)
		if err != nil && err != db.ErrNoMoreRows {
			suite.Fail("Did not replace user")
		}
	}
}

func TestSessionTestSuite(t *testing.T) {
	suite.Run(t, new(SessionTestSuite))
}
