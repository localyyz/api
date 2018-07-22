package apiclient

import (
	"context"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/web/auth"
)

type UserService service

func (c *UserService) Signup(ctx context.Context, user *data.User) (*auth.AuthUser, *http.Response, error) {
	postUserRequest := struct {
		Name            string          `json:"fullName,required"`
		Email           string          `json:"email,required"`
		Password        string          `json:"password,required"`
		PasswordConfirm string          `json:"passwordConfirm,required"`
		Gender          data.UserGender `json:"gender"`
	}{
		Name:            user.Name,
		Email:           user.Email,
		Password:        "test1234",
		PasswordConfirm: "test1234",
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
