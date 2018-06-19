package endtoend

import (
	"fmt"
	"net/http"
	"testing"

	"bitbucket.org/moodie-app/moodie-api/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SessionTestSuite struct {
	suite.Suite
	*fixture

	env *tests.Env
}

func (suite *SessionTestSuite) SetupSuite() {
	suite.env = tests.SetupEnv(suite.T())
	suite.fixture = &fixture{}
	suite.SetupData(suite.T())
}

func (suite *SessionTestSuite) TearDownSuite() {
	suite.TeardownData(suite.T())
}

func (suite *SessionTestSuite) TestSessionPublic() {
	// public routes can be accessed without session
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s", suite.env.URL), nil)
	// verify default cart is okay
	rr, _ := http.DefaultClient.Do(req)
	assert.Equal(suite.T(), http.StatusOK, rr.StatusCode)
}

func (suite *SessionTestSuite) TestSessionAnonymousWithDeviceID() {
	// anonymous user session with device id
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/carts/default", suite.env.URL), nil)
	req.Header.Add("X-DEVICE-ID", "test-device-token")

	// verify default cart is okay
	rr, _ := http.DefaultClient.Do(req)
	assert.Equal(suite.T(), http.StatusOK, rr.StatusCode)

}

func TestSessionTestSuite(t *testing.T) {
	suite.Run(t, new(SessionTestSuite))
}
