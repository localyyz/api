package endtoend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/tests"
	"bitbucket.org/moodie-app/moodie-api/web/auth"
	"github.com/stretchr/testify/suite"
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
	suite.SetupData(suite.T())
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
	suite.Equal(http.StatusUnauthorized, rr.StatusCode)
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
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(map[string]string{
		"fullName":        "newuser signup",
		"email":           "newuser@localyyz.com",
		"password":        "test1234",
		"passwordConfirm": "test1234",
	})

	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/signup", suite.env.URL), buf)

	// verify new user is created
	rr, _ := http.DefaultClient.Do(req)
	suite.Equal(http.StatusCreated, rr.StatusCode)

	var authUser *auth.AuthUser
	suite.NoError(json.NewDecoder(rr.Body).Decode(&authUser))

	// validate new user
	suite.Equal("newuser@localyyz.com", authUser.Username)
	suite.Equal("newuser signup", authUser.Name)
	suite.Equal("email", authUser.Network)
	suite.NotEmpty(authUser.ID)

	dbUser, err := data.DB.User.FindByID(authUser.ID)
	suite.NoError(err)
	suite.Nil(dbUser.DeviceToken)
}

func (suite *SessionTestSuite) TestSessionFacebookSignup() {
	connect.FacebookLogin = &MockFacebook{}

	buf := &bytes.Buffer{}
	json.NewEncoder(buf).Encode(map[string]string{
		"token":      "localyyz-test-token-signup",
		"inviteCode": "etc",
	})

	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/login/facebook", suite.env.URL), buf)
	uID := "test-device-token-email-signup"
	req.Header.Add("X-DEVICE-ID", uID)

	// verify new user is created
	rr, _ := http.DefaultClient.Do(req)
	suite.Equal(http.StatusOK, rr.StatusCode)

	var authUser *auth.AuthUser
	suite.NoError(json.NewDecoder(rr.Body).Decode(&authUser))


	//validate new user
	suite.Equal("facebook", authUser.Network)
	suite.Equal("test@localyyz.com", authUser.Email)

	//getting it from data to manually check
	user, _ :=data.DB.User.FindByID(authUser.ID)
	suite.Equal("test-device-token-email-signup", *user.DeviceToken)
	suite.NotEmpty(authUser.ID)
}

func TestSessionTestSuite(t *testing.T) {
	suite.Run(t, new(SessionTestSuite))
}
