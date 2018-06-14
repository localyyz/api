package cart

import (
	"context"
	"testing"

	"bitbucket.org/moodie-app/moodie-api/data"
)

var defaultUser = &data.User{Name: "Paul"}

type testUserOption func(*data.User) *data.User

func userWithEmail(email string) testUserOption {
	return func(u *data.User) *data.User {
		u.Email = email
		return u
	}
}

var defaultCart = &data.Cart{}

type testCartOption func(*data.Cart) *data.Cart

func cartWithAddress(hasBilling bool, hasShipping bool) testCartOption {
	return func(c *data.Cart) *data.Cart {
		if hasBilling {
			c.BillingAddress = &data.CartAddress{}
		}
		if hasShipping {
			c.ShippingAddress = &data.CartAddress{}
		}
		return c
	}
}

func TestValidateCheckout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		userOptions []testUserOption
		cartOptions []testCartOption
		expected    error
	}{
		{
			"valid checkout",
			[]testUserOption{userWithEmail("paul@localyyz")},
			[]testCartOption{cartWithAddress(true, true)},
			nil,
		},
		{
			"missing user email",
			[]testUserOption{},
			[]testCartOption{cartWithAddress(true, true)},
			ErrInvalidEmail,
		},
		{
			"missing shipping address",
			[]testUserOption{userWithEmail("paul@localyyz.com")},
			[]testCartOption{cartWithAddress(true, false)},
			ErrInvalidShipping,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := *defaultUser
			for _, opt := range tt.userOptions {
				opt(&u)
			}
			c := *defaultCart
			for _, opt := range tt.cartOptions {
				opt(&c)
			}

			ctx := context.WithValue(context.Background(), "session.user", &u)
			ctx = context.WithValue(ctx, "cart", &c)

			if actual := validateCart(ctx); actual != tt.expected {
				t.Errorf("test '%s': expected error '%v', got '%v'", tt.name, tt.expected, actual)
			}
		})
	}

}
