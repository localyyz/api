package apiclient

import (
	"context"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/web/auth"
)

type UserService service

func (c *UserService) SignupWithEmail(ctx context.Context, email, name, password string) (*auth.AuthUser, *http.Response, error) {
	postUserRequest := struct {
		Name            string          `json:"fullName,required"`
		Email           string          `json:"email,required"`
		Password        string          `json:"password,required"`
		PasswordConfirm string          `json:"passwordConfirm,required"`
		Gender          data.UserGender `json:"gender"`
	}{
		Name:            name,
		Email:           email,
		Password:        password,
		PasswordConfirm: password,
		// TODO: gender?
	}

	req, err := c.client.NewRequest("POST", "/signup", postUserRequest)
	if err != nil {
		return nil, nil, err
	}

	userResponse := new(auth.AuthUser)
	resp, err := c.client.Do(ctx, req, userResponse)
	if err != nil {
		return nil, resp, err
	}

	return userResponse, resp, nil
}

func (c *UserService) SignupWithFacebook(ctx context.Context, token string) (*auth.AuthUser, *http.Response, error) {
	postUserRequest := struct {
		Token string `json:"token"`
		InviteCode string `json:"inviteCode"`
	}{
		Token: token,
		InviteCode: "etc",
	}

	req, err := c.client.NewRequest("POST", "/login/facebook", postUserRequest )

	if err != nil {
		return nil, nil, err
	}

	userResponse := new(auth.AuthUser)
	resp, err := c.client.Do(ctx, req, userResponse)
	if err != nil {
		return nil, resp, err
	}
	return userResponse, resp, nil
}

func (c *UserService) LoginWithEmail(ctx context.Context, email, password string) (*auth.AuthUser, *http.Response, error) {
	postUserRequest := struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}{
		Email: email,
		Password: password,
	}

	req, err := c.client.NewRequest("POST", "/login", postUserRequest)

	if err != nil {
		return nil, nil, nil
	}

	userResponse := new(auth.AuthUser)
	resp, err := c.client.Do(ctx, req, userResponse)
	if err != nil {
		return nil, resp, err
	}
	return userResponse, resp, nil
}