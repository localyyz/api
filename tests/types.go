package tests

import (
	"fmt"
	"math/rand"
	"net/http"
	"testing"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/token"
	"bitbucket.org/moodie-app/moodie-api/web/auth"
	"github.com/go-chi/jwtauth"
	"github.com/stretchr/testify/assert"
)

type User struct {
	*auth.AuthUser
	client *http.Client
}

func NewAuthUser(t *testing.T) *User {
	n := rand.Int()

	// setup fixtures for test suite
	user := &data.User{
		Username:     fmt.Sprintf("user%d", n),
		Email:        "nobody@localyyz.com",
		Name:         "Paul X",
		Network:      "email",
		PasswordHash: string(""),
		LoggedIn:     true,
	}
	assert.NoError(t, data.DB.Save(user))
	token, _ := token.Encode(jwtauth.Claims{"user_id": user.ID})

	return &User{
		AuthUser: &auth.AuthUser{User: user, JWT: token.Raw},
		client:   http.DefaultClient,
	}
}
