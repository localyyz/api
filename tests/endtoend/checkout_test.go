package endtoend

import (
	"bitbucket.org/moodie-app/moodie-api/tests"
	"github.com/stretchr/testify/suite"
	"bitbucket.org/moodie-app/moodie-api/data"
	"net/http"
	"fmt"
	"github.com/stretchr/testify/assert"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"upper.io/db.v3"
	"github.com/stretchr/testify/require"
	"bitbucket.org/moodie-app/moodie-api/lib/shopper"
	"regexp"
	"testing"
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

func (suite *CheckoutTestSuite) TearDownSuite() {
	suite.TeardownData(suite.T())
}

// E2E checkout tests

func (suite *CheckoutTestSuite) TestCheckoutSuccess() {
	fullAddress := &data.CartAddress{
		Address:   "123 Toronto Street",
		FirstName: "User",
		LastName:  "Localyyz",
		City:      "Toronto",
		Country:   "Canada",
		Province:  "Ontario",
		Zip:       "M5J 1B7",
	}

	{ // verify default cart exists
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/carts/default", suite.env.URL), nil)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user1.JWT))

		// verify default cart is okay
		rr, _ := http.DefaultClient.Do(req)
		assert.Equal(suite.T(), http.StatusOK, rr.StatusCode)
	}

	{ // verify add to cart as user
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{"variantId": suite.variantInStock.ID})
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/default/items", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user1.JWT))

		// verify default cart is okay
		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusCreated, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}
	}

	{ // attempt to checkout, verify bad request
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/default/checkout", suite.env.URL), nil)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user1.JWT))

		// verify default cart is okay
		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusBadRequest, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}
	}

	{ // update cart addresses + email
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{
			"shippingAddress": fullAddress,
			"billingAddress":  fullAddress,
			"email":           suite.user1.Email,
		})
		req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/carts/default", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user1.JWT))

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
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user1.JWT))

		// verify default cart is okay
		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusOK, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}

		var cart *presenter.Cart
		json.NewDecoder(rr.Body).Decode(&cart)

		// validate cart
		if assert.NotNil(suite.T(), cart) {
			// assert that the cart returned has all the valid fields
			assert.Equal(suite.T(), fullAddress, cart.ShippingAddress.CartAddress)
			assert.Equal(suite.T(), fullAddress, cart.BillingAddress.CartAddress)
		}
		assert.NotNil(suite.T(), cart.CartItems)

		if assert.NotNil(suite.T(), cart.Checkouts) {
			// pricing
			assert.Equal(suite.T(), 213.60, cart.Checkouts[0].SubtotalPrice)
			assert.Equal(suite.T(), 251.84, cart.Checkouts[0].TotalPrice)
			assert.Equal(suite.T(), "251.84", cart.Checkouts[0].PaymentDue)

			// shipping line
			assert.NotNil(suite.T(), cart.Checkouts[0].ShippingLine)
			assert.Equal(suite.T(), 10.47, cart.Checkouts[0].TotalShipping)

			// tax lines
			assert.NotNil(suite.T(), cart.Checkouts[0].TaxLines)
			assert.Equal(suite.T(), 0.13, cart.Checkouts[0].TaxLines[0].Rate)
			assert.Equal(suite.T(), "27.77", cart.Checkouts[0].TaxLines[0].Price)
			assert.Equal(suite.T(), "HST", cart.Checkouts[0].TaxLines[0].Title)

			// ids
			assert.NotEmpty(suite.T(), cart.Checkouts[0].ID)
			assert.Equal(suite.T(), suite.user1.ID, cart.Checkouts[0].UserID)
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
				Name:   "User Localyyz",
				CVC:    "123",
			},
		})
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/default/pay", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user1.JWT))

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
		assert.Equal(suite.T(), data.CartStatusComplete, cart.Status)
		if assert.NotNil(suite.T(), cart.Checkouts) {
			assert.Equal(suite.T(), data.CheckoutStatusPaymentSuccess, cart.Checkouts[0].Status)

			dbCheckout, err := data.DB.Checkout.FindOne(db.Cond{"id": cart.Checkouts[0].ID})
			require.NoError(suite.T(), err)
			assert.NotEmpty(suite.T(), dbCheckout.SuccessPaymentID)
		}
	}
}

func (suite *CheckoutTestSuite) TestCheckoutNotInStock() {

	{ //add to cart
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{"variantId": suite.variantNotInStock.ID})
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/default/items", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user2.JWT))
		http.DefaultClient.Do(req)
	}

	{ // update cart addresses + email
		fullAddress := &data.CartAddress{
			Address:   "123 Toronto Street",
			FirstName: "User",
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
			"email":           suite.user2.Email,
		})
		req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/carts/default", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user2.JWT))
		http.DefaultClient.Do(req)
	}

	{ // attempt to checkout
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/default/checkout", suite.env.URL), nil)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user2.JWT))

		// verify it says there arent enough stock
		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusBadRequest, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}
		body, _ := ioutil.ReadAll(rr.Body)
		reg, _ := regexp.Compile("quantity Not enough items available")
		match := reg.MatchString(string(body))
		assert.Equal(suite.T(), true, match)
	}
}

func (suite *CheckoutTestSuite) TestCheckoutInvalidAddress() {

	{ // add to cart as user
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{"variantId": suite.variantInStock.ID})
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/default/items", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user3.JWT))
		http.DefaultClient.Do(req)
	}

	{ // update cart addresses + email
		fullAddress := &data.CartAddress{
			Address:   "123 Toronto Street",
			FirstName: "user",
			LastName:  "Localyyz",
			City:      "Toronto",
			Country:   "Canada",
			Province:  "Ontario",
			Zip:       "1234",
		}

		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{
			"shippingAddress": fullAddress,
			"billingAddress":  fullAddress,
			"email":           suite.user3.Email,
		})
		req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/carts/default", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user3.JWT))
		http.DefaultClient.Do(req)
	}

	{
		// attempt to checkout
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/default/checkout", suite.env.URL), nil)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user3.JWT))

		// verify it says zip is invalid
		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusBadRequest, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}
		body, _ := ioutil.ReadAll(rr.Body)
		reg, _ := regexp.Compile("zip is not valid for Canada")
		match := reg.MatchString(string(body))
		assert.Equal(suite.T(), true, match)
	}
}

func (suite *CheckoutTestSuite) TestCheckoutInvalidCVC() {
	{ //add to cart as user
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{"variantId": suite.variantInStock.ID})
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/default/items", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user4.JWT))

		http.DefaultClient.Do(req)
	}

	{ // update cart addresses + email
		fullAddress := &data.CartAddress{
			Address:   "123 Toronto Street",
			FirstName: "user",
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
			"email":           suite.user4.Email,
		})
		req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/carts/default", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user4.JWT))
		http.DefaultClient.Do(req)
	}

	{ // attempt to checkout
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/default/checkout", suite.env.URL), nil)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user4.JWT))
		http.DefaultClient.Do(req)
	}

	{ // pay.
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]*shopper.PaymentCard{
			"payment": {
				Number: "4000000000000127", //stripe credit card which will give invalid cvc
				Expiry: "12/22",
				Name:   "user Localyyz",
				CVC:    "123",
			},
		})
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/default/pay", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user4.JWT))

		// verify it returns cvc is invalid
		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusBadRequest, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}
		body, _ := ioutil.ReadAll(rr.Body)
		reg, _ := regexp.Compile("Your card's security code is incorrect")
		match := reg.MatchString(string(body))
		assert.Equal(suite.T(), true, match)
	}
}

func (suite *CheckoutTestSuite) TestCheckoutDoesNotShip() {
	{ //add to cart as user
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{"variantId": suite.variantInStock.ID})
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/default/items", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user5.JWT))

		http.DefaultClient.Do(req)
	}

	{ // update cart addresses + email
		fullAddress := &data.CartAddress{
			Address:   "123 London Street",
			FirstName: "user",
			LastName:  "Localyyz",
			City:      "London",
			Country:   "United Kindom",
			Province:  "Marylebone",
			Zip:       "W1U 8ED",
		}

		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{
			"shippingAddress": fullAddress,
			"billingAddress":  fullAddress,
			"email":           suite.user5.Email,
		})
		req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/carts/default", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user5.JWT))
		http.DefaultClient.Do(req)
	}

	{ // attempt to checkout
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/default/checkout", suite.env.URL), nil)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user5.JWT))
		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusBadRequest, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}
		b, _ := ioutil.ReadAll(rr.Body)
		assert.Contains(suite.T(), string(b), "shipping_address: country is not supported")
	}
}


func (suite *CheckoutTestSuite) TestCheckoutWithDiscountSuccess() {
	fullAddress := &data.CartAddress{
		Address:   "123 Toronto Street",
		FirstName: "Someone",
		LastName:  "Localyyz",
		City:      "Toronto",
		Country:   "Canada",
		Province:  "Ontario",
		Zip:       "M5J 1B7",
	}

	{ // verify add to cart as user
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{"variantId": suite.variantWithDiscount.ID})
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/default/items", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user6.JWT))

		// verify default cart is okay
		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusCreated, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}
	}

	var checkoutToValidate *data.Checkout
	{ // update cart addresses + email
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{
			"shippingAddress": fullAddress,
			"billingAddress":  fullAddress,
			"email":           suite.user6.Email,
		})
		req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/carts/default", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user6.JWT))

		// verify default cart is okay
		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusOK, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}

		var cart *presenter.Cart
		json.NewDecoder(rr.Body).Decode(&cart)
		assert.NotNil(suite.T(), cart)
		assert.NotEmpty(suite.T(), cart.Checkouts)
		checkoutToValidate = cart.Checkouts[0].Checkout
	}

	{ // update checkout with discount code
		discountCode := "TEST_SALE_CODE"
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{
			"discount": discountCode,
		})
		req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/carts/default/checkout/%d", suite.env.URL, checkoutToValidate.ID), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user6.JWT))

		// verify checkout is okay
		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusOK, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}

		var checkout *presenter.Checkout
		json.NewDecoder(rr.Body).Decode(&checkout)
		assert.NotNil(suite.T(), checkout)
		assert.Equal(suite.T(), discountCode, checkout.DiscountCode)
	}

	{ // checkout
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/default/checkout", suite.env.URL), nil)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user6.JWT))

		// verify default cart is okay
		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusOK, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}

		var cart *presenter.Cart
		json.NewDecoder(rr.Body).Decode(&cart)

		assert.NotNil(suite.T(), cart.CartItems)
		if assert.NotNil(suite.T(), cart.Checkouts) {
			// verify discount is applied
			if assert.NotNil(suite.T(), cart.Checkouts[0].AppliedDiscount.AppliedDiscount) {
				assert.NotEmpty(suite.T(), cart.Checkouts[0].AppliedDiscount.Amount)
				assert.Equal(suite.T(), "31.36", cart.Checkouts[0].AppliedDiscount.Amount)
				assert.NotEmpty(suite.T(), cart.Checkouts[0].AppliedDiscount.Title)
			}
		}
		assert.EqualValues(suite.T(), 3136, cart.TotalDiscount)
		assert.EqualValues(suite.T(), 1047, cart.TotalShipping)
		assert.EqualValues(suite.T(), 3669, cart.TotalTax)
		assert.EqualValues(suite.T(), 32940, cart.TotalPrice)
	}
}


func (suite *CheckoutTestSuite) TestCheckoutWithDiscountFailure() {
	fullAddress := &data.CartAddress{
		Address:   "123 Toronto Street",
		FirstName: "User",
		LastName:  "Localyyz",
		City:      "Toronto",
		Country:   "Canada",
		Province:  "Ontario",
		Zip:       "M5J 1B7",
	}
	u := suite.user7

	{ // verify add to cart as user
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{"variantId": suite.variantWithDiscount.ID})
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/default/items", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", u.JWT))

		// verify default cart is okay
		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusCreated, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}
	}

	var checkoutToValidate *data.Checkout
	{ // update cart addresses + email
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{
			"shippingAddress": fullAddress,
			"billingAddress":  fullAddress,
			"email":           u.Email,
		})
		req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/carts/default", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", u.JWT))

		// verify default cart is okay
		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusOK, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}

		var cart *presenter.Cart
		json.NewDecoder(rr.Body).Decode(&cart)
		assert.NotNil(suite.T(), cart)
		assert.NotEmpty(suite.T(), cart.Checkouts)
		checkoutToValidate = cart.Checkouts[0].Checkout
	}

	{ // update checkout with discount code
		discountCode := "INVALID_DISCOUNT_CODE"
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{
			"discount": discountCode,
		})
		req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/carts/default/checkout/%d", suite.env.URL, checkoutToValidate.ID), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", u.JWT))

		// verify checkout is okay
		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusOK, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}

		var checkout *presenter.Checkout
		json.NewDecoder(rr.Body).Decode(&checkout)
		assert.NotNil(suite.T(), checkout)
		assert.Equal(suite.T(), discountCode, checkout.DiscountCode)
	}

	{ // checkout
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/default/checkout", suite.env.URL), nil)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", u.JWT))

		// verify
		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusBadRequest, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}
		var cart *presenter.Cart
		json.NewDecoder(rr.Body).Decode(&cart)
		assert.True(suite.T(), cart.HasError)
		assert.Contains(suite.T(), cart.Error, "Unable to find a valid discount matching the code entered")
	}
}


func (suite *CheckoutTestSuite) TestHappyExpressCheckout() {

	{ // verify default cart exists
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/carts/express", suite.env.URL), nil)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user8.JWT))

		// verify default cart is okay
		rr, _ := http.DefaultClient.Do(req)
		assert.Equal(suite.T(), http.StatusOK, rr.StatusCode)
	}

	{ // verify add to cart as user
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{"productId": suite.variantInStock.ProductID, "color": "deep", "size": "small", "quantity": 1})
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/express/items", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user8.JWT))

		// verify default cart is okay
		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusOK, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}
	}

	{
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{
			"FirstName":    "Test",
			"LastName":     "Test",
			"Address":      "12 Deerford Road",
			"AddressOpt":   "",
			"City":         "Toronto",
			"Country":      "Canada",
			"CountryCode":  "CA",
			"Province":     "Ontario",
			"ProvinceCode": "ON",
			"Zip":          "M2J3J3",
			"isPartial": false,
		})

		req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/carts/express/shipping/address", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user8.JWT))

		// verify default cart is okay
		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusOK, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}
	}

	{
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{
			"Handle": "canada_post-DOM.EP-10.47",
		})

		req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/carts/express/shipping/method", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user8.JWT))

		// verify default cart is okay
		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusOK, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}
	}

	{
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/carts/express/shipping/estimate", suite.env.URL), nil)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user8.JWT))

		// verify default cart is okay
		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusOK, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}
	}

	{ // pay.
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{
			"BillingAddress": &data.CartAddress{
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
			"Email":               "waseef@localyyz.com",
			"ExpressPaymentToken": "tok_ca",
		})
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/express/pay", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user8.JWT))

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
		assert.Equal(suite.T(), data.CartStatusPaymentSuccess, cart.Status)

	}

}

func (suite *CheckoutTestSuite) TestExpiredLightningCollectionCheckout(){
	{ // verify default cart exists
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/carts/express", suite.env.URL), nil)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user9.JWT))
	}
	{ // verify add to cart as user
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{"productId": suite.variantLightningExpired.ProductID, "color": "deep", "size": "small", "quantity": 1})
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/express/items", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user9.JWT))

		// verify default cart is okay
		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusBadRequest, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}
	}
}

func (suite *CheckoutTestSuite) TestCapHitLightningCollectionCheckout(){
	{ // verify default cart exists
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/carts/express", suite.env.URL), nil)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user9.JWT))
	}
	{ // verify add to cart as user
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{"productId": suite.variantLightningCapHit.ProductID, "color": "deep", "size": "small", "quantity": 1})
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/express/items", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user9.JWT))

		// verify default cart is okay
		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusBadRequest, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}
	}
}


func (suite *CheckoutTestSuite) TestValidLightningCollectionCheckout(){
	{ // verify default cart exists
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/carts/express", suite.env.URL), nil)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user10.JWT))
	}
	{ // verify add to cart as user
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{"productId": suite.variantLightningValid.ProductID, "color": "deep", "size": "small", "quantity": 1})
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/express/items", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user10.JWT))

		// verify default cart is okay
		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusOK, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}
	}

	{
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{
			"FirstName":    "Test",
			"LastName":     "Test",
			"Address":      "12 Deerford Road",
			"AddressOpt":   "",
			"City":         "Toronto",
			"Country":      "Canada",
			"CountryCode":  "CA",
			"Province":     "Ontario",
			"ProvinceCode": "ON",
			"Zip":          "M2J3J3",
			"isPartial": true,
		})

		req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/carts/express/shipping/address", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user10.JWT))

		// verify default cart is okay
		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusOK, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}
	}

	{
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{
			"Handle": "canada_post-DOM.EP-10.47",
		})

		req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/carts/express/shipping/method", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user10.JWT))

		// verify default cart is okay
		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusOK, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}
	}

	{
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/carts/express/shipping/estimate", suite.env.URL), nil)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user10.JWT))

		// verify default cart is okay
		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusOK, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}
	}

	{ // pay.
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{
			"BillingAddress": &data.CartAddress{
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
			"Email":               "waseef@localyyz.com",
			"ExpressPaymentToken": "tok_ca",
		})
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/express/pay", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user10.JWT))

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
		assert.Equal(suite.T(), data.CartStatusPaymentSuccess, cart.Status)
	}

}


func TestCheckoutTestSuite(t *testing.T) {
	suite.Run(t, new(CheckoutTestSuite))
}
