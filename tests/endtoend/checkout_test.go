package endtoend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/shopper"
	"bitbucket.org/moodie-app/moodie-api/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	db "upper.io/db.v3"
)

type CheckoutTestSuite struct {
	suite.Suite
	*fixture

	env *tests.Env
}

func (suite *CheckoutTestSuite) SetupSuite() {
	suite.env = tests.SetupEnv(suite.T())
	suite.fixture = &fixture{}
	suite.SetupData(suite.T())
}

// E2E checkout tests
func (suite *CheckoutTestSuite) TestCheckoutSuccess() {
	{ // verify default cart exists
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/carts/default", suite.env.URL), nil)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.paul.JWT))

		// verify default cart is okay
		rr, _ := http.DefaultClient.Do(req)
		assert.Equal(suite.T(), http.StatusOK, rr.StatusCode)
	}

	{ // verify add to cart as user
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{"variantId": suite.variant1.ID})
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/default/items", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.paul.JWT))

		// verify default cart is okay
		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusCreated, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}
	}

	{ // attempt to checkout, verify bad request
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/default/checkout", suite.env.URL), nil)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.paul.JWT))

		// verify default cart is okay
		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusBadRequest, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}
	}

	{ // update cart addresses + email
		fullAddress := &data.CartAddress{
			Address:   "123 Toronto Street",
			FirstName: "Paul",
			LastName:  "Localyyz",
			City:      "Toronto",
			Country:   "Canada",
			Province:  "Ontario",
			Zip:       "M5J 1B7",
		}

		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{
			"shippingAddress": fullAddress,
			"billingAddress":  fullAddress,
			"email":           suite.paul.Email,
		})
		req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/carts/default", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.paul.JWT))

		// verify default cart is okay
		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusOK, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}

		var cart *presenter.Cart
		json.NewDecoder(rr.Body).Decode(&cart)
		assert.NotNil(suite.T(), cart)
		assert.NotNil(suite.T(), cart.ShippingAddress)
		assert.NotNil(suite.T(), cart.BillingAddress)
		assert.NotEmpty(suite.T(), cart.Email)
	}

	{ // reattempt to checkout
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/default/checkout", suite.env.URL), nil)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.paul.JWT))

		// verify default cart is okay
		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusOK, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}

		var cart *presenter.Cart
		json.NewDecoder(rr.Body).Decode(&cart)

		// validate cart
		assert.NotNil(suite.T(), cart)
		assert.NotNil(suite.T(), cart.CartItems)

		if assert.NotNil(suite.T(), cart.Checkouts) {
			// pricing
			assert.NotEmpty(suite.T(), cart.Checkouts[0].SubtotalPrice)
			assert.NotEmpty(suite.T(), cart.Checkouts[0].TotalPrice)
			assert.NotEmpty(suite.T(), cart.Checkouts[0].PaymentDue)

			// shipping line
			assert.NotNil(suite.T(), cart.Checkouts[0].ShippingLine)
			assert.Equal(suite.T(), 10.00, cart.Checkouts[0].TotalShipping)

			// tax lines
			assert.NotNil(suite.T(), cart.Checkouts[0].TaxLines)
			assert.Equal(suite.T(), 0.13, cart.Checkouts[0].TaxLines[0].Rate)
			assert.Equal(suite.T(), "HST", cart.Checkouts[0].TaxLines[0].Title)

			// ids
			assert.NotEmpty(suite.T(), cart.Checkouts[0].ID)
			assert.Equal(suite.T(), suite.paul.ID, cart.Checkouts[0].UserID)
			assert.Equal(suite.T(), cart.ID, *cart.Checkouts[0].CartID)

			dbCheckout, err := data.DB.Checkout.FindOne(db.Cond{"id": cart.Checkouts[0].ID})
			require.NoError(suite.T(), err)

			assert.NotEmpty(suite.T(), dbCheckout.Token)
			assert.NotEmpty(suite.T(), dbCheckout.PaymentAccountID)
		}
	}

	{ // pay.
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]*shopper.PaymentCard{
			"payment": {
				Number: "4242424242424242",
				Expiry: "12/22",
				Name:   "Paul Localyyz",
				CVC:    "123",
			},
		})
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/default/pay", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.paul.JWT))

		// verify default cart is okay
		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusOK, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}

		var cart *presenter.Cart
		json.NewDecoder(rr.Body).Decode(&cart)

		// validate cart
		assert.NotNil(suite.T(), cart)
		if assert.NotNil(suite.T(), cart.Checkouts) {
			assert.Equal(suite.T(), data.CheckoutStatusPaymentSuccess, cart.Checkouts[0].Status)

			dbCheckout, err := data.DB.Checkout.FindOne(db.Cond{"id": cart.Checkouts[0].ID})
			require.NoError(suite.T(), err)
			assert.NotEmpty(suite.T(), dbCheckout.SuccessPaymentID)
		}
	}

}

func TestCheckoutTestSuite(t *testing.T) {
	suite.Run(t, new(CheckoutTestSuite))
}
