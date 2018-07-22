package deals

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DealsTestSuite struct {
	suite.Suite
	*fixture

	env *tests.Env
}

type LightningError struct {
	Status string
	Error  string
}

func (suite *DealsTestSuite) SetupSuite() {
	suite.env = tests.SetupEnv(suite.T())
	suite.fixture = &fixture{}
	suite.SetupData(suite.T())
}

func (suite *DealsTestSuite) TearDownSuite() {
	suite.TeardownData(suite.T())
}

// E2E deals tests

func (suite *DealsTestSuite) TestValidLightningCollectionCheckout() {
	{ //verify default cart exists
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/carts/express", suite.env.URL), nil)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user9.JWT))

		rr, _ := http.DefaultClient.Do(req)
		assert.Equal(suite.T(), http.StatusOK, rr.StatusCode)
	}
	{ //add item to cart
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{"productId": suite.variantLightningValid.ProductID, "color": "deep", "size": "small", "quantity": 1})

		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/express/items", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user9.JWT))

		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusOK, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}
	}
	{ //put in address
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{
			"FirstName":    "Test",
			"LastName":     "Test",
			"Address":      "180 John Street",
			"AddressOpt":   "",
			"City":         "Toronto",
			"Country":      "Canada",
			"CountryCode":  "CA",
			"Province":     "Ontario",
			"ProvinceCode": "ON",
			"Zip":          "M2J3J3",
			"isPartial":    true,
		})

		req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/carts/express/shipping/address", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user9.JWT))

		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusOK, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}
	}
	{ //put in shipping method
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{
			"Handle": "canada_post-DOM.EP-10.47",
		})

		req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/carts/express/shipping/method", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user9.JWT))

		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusOK, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}
	}
	{ //get the shipping estimate
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/carts/express/shipping/estimate", suite.env.URL), nil)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user9.JWT))

		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusOK, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}
	}
	{ //pay
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{
			"BillingAddress": &data.CartAddress{
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
			"Email":               "waseef@localyyz.com",
			"ExpressPaymentToken": "tok_ca", //test token
		})

		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/express/pay", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user9.JWT))

		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusOK, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}

		var cart *presenter.Cart
		json.NewDecoder(rr.Body).Decode(&cart)

		assert.NotNil(suite.T(), cart)
		assert.Equal(suite.T(), data.CartStatusPaymentSuccess, cart.Status)
	}
}

func (suite *DealsTestSuite) TestExpiredLightningCollectionCheckout() {
	{ // verify default cart exists
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/carts/express", suite.env.URL), nil)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user10.JWT))

		rr, _ := http.DefaultClient.Do(req)
		assert.Equal(suite.T(), http.StatusOK, rr.StatusCode)
	}
	{ // verify add to cart as user
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{"productId": suite.variantLightningExpired.ProductID, "color": "deep", "size": "small", "quantity": 1})
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/express/items", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user10.JWT))

		// verify it returns bad request
		rr, _ := http.DefaultClient.Do(req)
		if assert.Equal(suite.T(), http.StatusBadRequest, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			var error LightningError
			json.Unmarshal([]byte(b), &error)
			assert.Equal(suite.T(), error.Error, "this lightning deal has ended")
		} else {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}
	}
}

func (suite *DealsTestSuite) TestCapHitLightningCollectionCheckout() {
	{ //add item to cart
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{"productId": suite.variantLightningCapHit.ProductID, "color": "deep", "size": "small", "quantity": 1})

		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/express/items", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user11.JWT))
		http.DefaultClient.Do(req)
	}
	{ //put in address
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{
			"FirstName":    "Test",
			"LastName":     "Test",
			"Address":      "180 John Street",
			"AddressOpt":   "",
			"City":         "Toronto",
			"Country":      "Canada",
			"CountryCode":  "CA",
			"Province":     "Ontario",
			"ProvinceCode": "ON",
			"Zip":          "M2J3J3",
			"isPartial":    true,
		})

		req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/carts/express/shipping/address", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user11.JWT))

		http.DefaultClient.Do(req)
	}
	{ //put in shipping method
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{
			"Handle": "canada_post-DOM.EP-10.47",
		})

		req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/carts/express/shipping/method", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user11.JWT))

		http.DefaultClient.Do(req)
	}
	{ //get the shipping estimate
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/carts/express/shipping/estimate", suite.env.URL), nil)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user11.JWT))

		http.DefaultClient.Do(req)
	}
	{ //pay
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{
			"BillingAddress": &data.CartAddress{
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
			"Email":               "waseef@localyyz.com",
			"ExpressPaymentToken": "tok_ca", //test token
		})

		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/express/pay", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user11.JWT))
		http.DefaultClient.Do(req)
	}
	{ //try to add item to cart
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{"productId": suite.variantLightningCapHit.ProductID, "color": "deep", "size": "small", "quantity": 1})

		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/express/items", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user12.JWT))
		rr, _ := http.DefaultClient.Do(req)
		if assert.Equal(suite.T(), http.StatusBadRequest, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			var error LightningError
			json.Unmarshal([]byte(b), &error)
			assert.Equal(suite.T(), error.Error, "the products from this lightning collection have been sold out")
		} else {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}

	}
}

func (suite *DealsTestSuite) TestUserLimit() {
	{ //add item to cart
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{"productId": suite.variantMultiplePurchase.ProductID, "color": "deep", "size": "small", "quantity": 1})

		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/express/items", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user12.JWT))
		http.DefaultClient.Do(req)
	}
	{ //put in address
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{
			"FirstName":    "Test",
			"LastName":     "Test",
			"Address":      "180 John Street",
			"AddressOpt":   "",
			"City":         "Toronto",
			"Country":      "Canada",
			"CountryCode":  "CA",
			"Province":     "Ontario",
			"ProvinceCode": "ON",
			"Zip":          "M2J3J3",
			"isPartial":    true,
		})

		req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/carts/express/shipping/address", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user12.JWT))

		http.DefaultClient.Do(req)
	}
	{ //put in shipping method
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{
			"Handle": "canada_post-DOM.EP-10.47",
		})

		req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/carts/express/shipping/method", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user12.JWT))

		http.DefaultClient.Do(req)
	}
	{ //get the shipping estimate
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/carts/express/shipping/estimate", suite.env.URL), nil)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user12.JWT))

		http.DefaultClient.Do(req)
	}
	{ //pay
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{
			"BillingAddress": &data.CartAddress{
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
			"Email":               "waseef@localyyz.com",
			"ExpressPaymentToken": "tok_ca", //test token
		})

		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/express/pay", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user12.JWT))
		http.DefaultClient.Do(req)
	}
	{ //try to add item to cart
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{"productId": suite.variantMultiplePurchase.ProductID, "color": "deep", "size": "small", "quantity": 1})

		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/express/items", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user12.JWT))
		rr, _ := http.DefaultClient.Do(req)
		if assert.Equal(suite.T(), http.StatusBadRequest, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			var error LightningError
			json.Unmarshal([]byte(b), &error)
			assert.Equal(suite.T(), error.Error, "you have already purchased today's deal")
		} else {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}

	}
}

// test successfully activating an user deal, and checking out the deal
func (suite *DealsTestSuite) TestUserDealSuccess() {
	{ //verify default cart exists
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/carts/express", suite.env.URL), nil)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user9.JWT))

		rr, _ := http.DefaultClient.Do(req)
		assert.Equal(suite.T(), http.StatusOK, rr.StatusCode)
	}
	{ //add item to cart
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{"productId": suite.variantLightningValid.ProductID, "color": "deep", "size": "small", "quantity": 1})

		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/express/items", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user9.JWT))

		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusOK, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}
	}
	{ //put in address
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{
			"FirstName":    "Test",
			"LastName":     "Test",
			"Address":      "180 John Street",
			"AddressOpt":   "",
			"City":         "Toronto",
			"Country":      "Canada",
			"CountryCode":  "CA",
			"Province":     "Ontario",
			"ProvinceCode": "ON",
			"Zip":          "M2J3J3",
			"isPartial":    true,
		})

		req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/carts/express/shipping/address", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user9.JWT))

		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusOK, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}
	}
	{ //put in shipping method
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{
			"Handle": "canada_post-DOM.EP-10.47",
		})

		req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/carts/express/shipping/method", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user9.JWT))

		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusOK, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}
	}
	{ //get the shipping estimate
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/carts/express/shipping/estimate", suite.env.URL), nil)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user9.JWT))

		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusOK, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}
	}
	{ //pay
		b := &bytes.Buffer{}
		json.NewEncoder(b).Encode(map[string]interface{}{
			"BillingAddress": &data.CartAddress{
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
			"Email":               "waseef@localyyz.com",
			"ExpressPaymentToken": "tok_ca", //test token
		})

		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/carts/express/pay", suite.env.URL), b)
		req.Header.Add("Authorization", fmt.Sprintf("BEARER %s", suite.user9.JWT))

		rr, _ := http.DefaultClient.Do(req)
		if !assert.Equal(suite.T(), http.StatusOK, rr.StatusCode) {
			b, _ := ioutil.ReadAll(rr.Body)
			assert.FailNow(suite.T(), string(b))
		}

		var cart *presenter.Cart
		json.NewDecoder(rr.Body).Decode(&cart)

		assert.NotNil(suite.T(), cart)
		assert.Equal(suite.T(), data.CartStatusPaymentSuccess, cart.Status)
	}
}

func TestDealsTestSuite(t *testing.T) {
	suite.Run(t, new(DealsTestSuite))
}
