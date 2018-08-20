package endtoend

import (
	"context"
	"net/http"
	"testing"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopper"
	"bitbucket.org/moodie-app/moodie-api/tests"
	"bitbucket.org/moodie-app/moodie-api/tests/apiclient"
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
	suite.SetupData(suite.T(), suite.env.URL)
}

func (suite *CheckoutTestSuite) TearDownSuite() {
	suite.TeardownData(suite.T())
}

func (suite *CheckoutTestSuite) TearDownTest() {
	data.DB.Exec("TRUNCATE carts cascade;")
}

// E2E checkout tests

func (suite *CheckoutTestSuite) TestCheckoutSuccess() {
	user := suite.user
	client := user.client
	ctx := context.Background()

	_, resp, err := client.Cart.AddItem(ctx, &data.CartItem{VariantID: suite.variantInStock.ID})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), http.StatusCreated, resp.StatusCode)

	fullAddress := &data.CartAddress{
		Address:   "123 Toronto Street",
		FirstName: "Someone",
		LastName:  "Localyyz",
		City:      "Toronto",
		Country:   "Canada",
		Province:  "Ontario",
		Zip:       "M5J 1B7",
	}
	cart, _, err := client.Cart.Put(ctx, &data.Cart{
		ShippingAddress: fullAddress,
		BillingAddress:  fullAddress,
		Email:           user.Email,
	})
	require.NoError(suite.T(), err)

	suite.NotNil(cart)
	suite.NotNil(cart.ShippingAddress)
	suite.NotNil(cart.BillingAddress)
	suite.NotEmpty(cart.Email)

	// attempt to checkout
	cart, _, err = client.Cart.Checkout(ctx)
	require.NotNil(suite.T(), cart)
	// assert that the cart returned has all the valid fields
	suite.Equal(fullAddress, cart.ShippingAddress.CartAddress)
	suite.Equal(fullAddress, cart.BillingAddress.CartAddress)
	suite.NotEmpty(cart.CartItems)
	require.NotEmpty(suite.T(), cart.Checkouts)

	// pricing
	suite.InDelta(213.60, cart.Checkouts[0].SubtotalPrice, 0.1)
	suite.InDelta(251.98, cart.Checkouts[0].TotalPrice, 0.1)
	suite.Equal("251.98", cart.Checkouts[0].PaymentDue)

	// shipping line
	suite.NotNil(cart.Checkouts[0].ShippingLine)
	suite.InDelta(10.61, cart.Checkouts[0].TotalShipping, 0.1)

	// tax lines
	suite.NotNil(cart.Checkouts[0].TaxLines)
	suite.InDelta(0.13, cart.Checkouts[0].TaxLines[0].Rate, 0.1)
	suite.Equal("27.77", cart.Checkouts[0].TaxLines[0].Price)
	suite.Equal("HST", cart.Checkouts[0].TaxLines[0].Title)

	// ids
	suite.NotEmpty(cart.Checkouts[0].ID)
	suite.Equal(suite.user.ID, cart.Checkouts[0].UserID)
	suite.Equal(cart.ID, *cart.Checkouts[0].CartID)

	dbCheckout, err := data.DB.Checkout.FindOne(db.Cond{"id": cart.Checkouts[0].ID})
	suite.NoError(err)

	suite.NotEmpty(dbCheckout.Token)
	suite.NotEmpty(dbCheckout.PaymentAccountID)

	{ // pay.
		ctx := context.Background()
		paymentCard := &shopper.PaymentCard{
			Number: "4242424242424242",
			Expiry: "12/28",
			Name:   "Someone Localyyz",
			CVC:    "123",
		}
		cart, resp, err := client.Cart.Pay(ctx, paymentCard)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), http.StatusOK, resp.StatusCode)

		// validate cart
		suite.NotNil(cart)
		suite.Equal(data.CartStatusComplete, cart.Status)
		if !suite.NotEmpty(cart.Checkouts) {
			suite.FailNow("unexpected empty checkout")
		}

		suite.Equal(data.CheckoutStatusPaymentSuccess, cart.Checkouts[0].Status)
		dbCheckout, err := data.DB.Checkout.FindOne(db.Cond{"id": cart.Checkouts[0].ID})
		require.NoError(suite.T(), err)
		assert.NotEmpty(suite.T(), dbCheckout.SuccessPaymentID)
	}
}

// TEST: checkout on item not in stock
func (suite *CheckoutTestSuite) TestCheckoutNotInStock() {
	user := suite.user
	client := user.client

	_, _, err := client.Cart.AddItem(
		context.Background(),
		&data.CartItem{
			VariantID: suite.variantNotInStock.ID,
		},
	)
	require.NotNil(suite.T(), err)

	apiErr := err.(apiclient.Err400)
	suite.Contains(apiErr.Message, "variant is out of stock")
}

func (suite *CheckoutTestSuite) TestCheckoutNotInStockRemotely() {
	user := suite.user
	client := user.client

	ctx := context.Background()
	_, _, err := client.Cart.AddItem(
		ctx,
		&data.CartItem{
			VariantID: suite.variantNotInStockRemotely.ID,
		},
	)
	require.NoError(suite.T(), err)

	_, _, err = client.Cart.Put(ctx, &data.Cart{
		ShippingAddress: suite.validAddress,
		BillingAddress:  suite.validAddress,
		Email:           user.Email,
	})
	suite.NoError(err)

	cart, _, _ := client.Cart.Checkout(context.Background())
	if suite.NotEmpty(cart.CartItems) {
		suite.True(cart.CartItems[0].HasError)
		suite.Contains(cart.CartItems[0].Err, "out of stock")
	}
}

// TEST invalid "zip code" for Canada
func (suite *CheckoutTestSuite) TestCheckoutInvalidAddress() {
	user := suite.user
	client := user.client

	{
		// add to cart as user
		ctx := context.Background()
		_, _, err := client.Cart.AddItem(
			ctx,
			&data.CartItem{
				VariantID: suite.variantInStock.ID,
			},
		)
		suite.NoError(err)

		// update cart shipping/billing addresses
		invalidAddress := &data.CartAddress{
			Address:   "123 Toronto Street",
			FirstName: "user",
			LastName:  "Localyyz",
			City:      "Toronto",
			Country:   "Canada",
			Province:  "Ontario",
			Zip:       "1234",
		}
		_, _, err = client.Cart.Put(ctx, &data.Cart{
			ShippingAddress: invalidAddress,
			BillingAddress:  suite.validAddress,
			Email:           user.Email,
		})
		suite.NoError(err)
	}

	{
		cart, _, err := client.Cart.Checkout(context.Background())
		require.Nil(suite.T(), err, "unexpected checkout error %v", err)
		require.True(suite.T(), cart.HasError)
		require.True(suite.T(), cart.ShippingAddress.HasError)
		suite.Contains(cart.ShippingAddress.Error, "zip is not valid for Canada")
		suite.Contains(cart.Error, "zip is not valid for Canada")
	}
}

// TEST: invalid CVC number provided
func (suite *CheckoutTestSuite) TestCheckoutInvalidCVC() {
	user := suite.user
	client := user.client

	{
		// add to cart as user
		ctx := context.Background()
		_, _, err := client.Cart.AddItem(
			ctx,
			&data.CartItem{
				VariantID: suite.variantInStock.ID,
			},
		)
		suite.NoError(err)

		// update cart shipping/billing addresses
		fullAddress := &data.CartAddress{
			Address:   "123 Toronto Street",
			FirstName: "user",
			LastName:  "Localyyz",
			City:      "Toronto",
			Country:   "Canada",
			Province:  "Ontario",
			Zip:       "M5J 1B7",
		}
		_, _, err = client.Cart.Put(ctx, &data.Cart{
			ShippingAddress: fullAddress,
			BillingAddress:  fullAddress,
			Email:           user.Email,
		})
		suite.NoError(err)

		// create checkout
		_, _, err = client.Cart.Checkout(context.Background())
		suite.NoError(err)
	}

	{ // pay.
		ctx := context.Background()
		paymentCard := &shopper.PaymentCard{
			Number: "4000000000000127", //stripe credit card which will give invalid cvc
			Expiry: "12/22",
			Name:   "user Localyyz",
			CVC:    "123",
		}
		_, _, err := client.Cart.Pay(ctx, paymentCard)
		require.NotNil(suite.T(), err)
		apiErr, ok := err.(apiclient.Err400)
		if !ok {
			suite.FailNow("unknown error", "expected api error got: %+v", err)
		}
		suite.Contains(apiErr.Message, "Security code was not matched by the processor")
	}
}

// TEST: does not ship to address error
func (suite *CheckoutTestSuite) TestCheckoutDoesNotShip() {
	user := suite.user
	client := user.client

	{
		// add to cart as user
		ctx := context.Background()
		_, _, err := client.Cart.AddItem(
			ctx,
			&data.CartItem{
				VariantID: suite.variantInStock.ID,
			},
		)
		suite.NoError(err)

		// update cart shipping/billing addresses
		dnsAddress := &data.CartAddress{
			Address:   "123 London Street",
			FirstName: "user",
			LastName:  "Localyyz",
			City:      "London",
			Country:   "United Kindom",
			Province:  "Marylebone",
			Zip:       "W1U 8ED",
		}
		_, _, err = client.Cart.Put(ctx, &data.Cart{
			ShippingAddress: dnsAddress,
			BillingAddress:  suite.validAddress,
			Email:           user.Email,
		})
		suite.NoError(err)
	}

	cart, _, err := client.Cart.Checkout(context.Background())
	require.Nil(suite.T(), err, "unexpected checkout error %v", err)
	require.NotNil(suite.T(), cart)
	require.True(suite.T(), cart.HasError)
	suite.Contains(cart.Error, "country is not supported")
	require.True(suite.T(), cart.ShippingAddress.HasError)
	suite.Contains(cart.ShippingAddress.Error, "country is not supported")
}

//TEST: apply valid discount
func (suite *CheckoutTestSuite) TestCheckoutWithDiscountSuccess() {
	user := suite.user
	client := user.client

	{
		// add to cart as user
		ctx := context.Background()
		_, _, err := client.Cart.AddItem(
			ctx,
			&data.CartItem{
				VariantID: suite.variantWithDiscount.ID,
			},
		)
		suite.NoError(err)

		cart, _, err := client.Cart.Put(ctx, &data.Cart{
			ShippingAddress: suite.validAddress,
			BillingAddress:  suite.validAddress,
			Email:           user.Email,
		})
		if suite.NoError(err) {
			checkoutToValidate := cart.Checkouts[0]

			discountCode := "TEST_SALE_CODE"
			ctx := context.Background()
			checkout, _, err := client.Cart.AddDiscountCode(ctx, checkoutToValidate.ID, discountCode)

			suite.NoError(err)
			// validate that discount code is saved/applied
			suite.Equal(discountCode, checkout.DiscountCode)
		}
	}

	{ // checkout
		ctx := context.Background()
		cart, _, err := client.Cart.Checkout(ctx)
		if suite.NoError(err) {
			suite.NotNil(cart.CartItems)

			if suite.NotNil(cart.Checkouts) {
				// verify discount is applied
				if suite.NotNil(cart.Checkouts[0].AppliedDiscount.AppliedDiscount) {
					suite.NotEmpty(cart.Checkouts[0].AppliedDiscount.Amount)
					suite.Equal("31.36", cart.Checkouts[0].AppliedDiscount.Amount)
					suite.NotEmpty(cart.Checkouts[0].AppliedDiscount.Title)
				}
			}
			suite.EqualValues(3136, cart.TotalDiscount)
			suite.EqualValues(1057, cart.TotalShipping)
			suite.EqualValues(3669, cart.TotalTax)
			suite.EqualValues(32950, cart.TotalPrice)
		}
	}
}

func (suite *CheckoutTestSuite) TestCheckoutWithDiscountFailure() {
	user := suite.user
	client := user.client

	{
		// add to cart as user
		ctx := context.Background()
		_, _, err := client.Cart.AddItem(
			ctx,
			&data.CartItem{
				VariantID: suite.variantWithDiscount.ID,
			},
		)
		suite.NoError(err)

		cart, _, err := client.Cart.Put(ctx, &data.Cart{
			ShippingAddress: suite.validAddress,
			BillingAddress:  suite.validAddress,
			Email:           user.Email,
		})
		if suite.NoError(err) {
			checkoutToValidate := cart.Checkouts[0]

			discountCode := "INVALID_DISCOUNT_CODE"
			ctx := context.Background()
			checkout, _, err := client.Cart.AddDiscountCode(ctx, checkoutToValidate.ID, discountCode)

			suite.NoError(err)
			// validate that discount code is saved/applied
			suite.Equal(discountCode, checkout.DiscountCode)
		}
	}

	{ // checkout
		ctx := context.Background()
		cart, _, err := client.Cart.Checkout(ctx)
		require.Nil(suite.T(), err, "unexpected cart error %+v", err)
		require.NotNil(suite.T(), cart)
		suite.Contains(cart.Error, "Unable to find a valid discount matching the code entered")
	}
}

func TestCheckoutTestSuite(t *testing.T) {
	suite.Run(t, new(CheckoutTestSuite))
}
