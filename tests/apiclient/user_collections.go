package apiclient

import (
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/user"
	"context"
	"fmt"
	"net/http"
)

type UserCollService service

func (s *UserCollService) CreateNewUserColl(ctx context.Context, title string) (*presenter.CollectionUser, *http.Response, error) {
	payload := &user.NewUserCollection{
		Title: title,
	}

	req, err := s.client.NewRequest("POST", "/users/collections/", payload)
	if err != nil {
		return nil, nil, err
	}

	userCollResponse := new(presenter.CollectionUser)
	resp, err := s.client.Do(ctx, req, userCollResponse)
	if err != nil {
		return nil, resp, err
	}

	return userCollResponse, resp, nil
}

func (s *UserCollService) ListUserColl(ctx context.Context) ([]*presenter.CollectionUser, *http.Response, error) {
	req, err := s.client.NewRequest("GET", "/users/collections/", nil)
	if err != nil {
		return nil, nil, err
	}

	var userCollResponse []*presenter.CollectionUser
	resp, err := s.client.Do(ctx, req, &userCollResponse)
	if err != nil {
		return nil, resp, err
	}

	return userCollResponse, resp, nil
}

func (s *UserCollService) GetUserColl(ctx context.Context, collectionID int64) (*presenter.CollectionUser, *http.Response, error) {
	req, err := s.client.NewRequest("GET", fmt.Sprintf("/users/collections/%d/", collectionID), nil)
	if err != nil {
		return nil, nil, err
	}

	userCollResponse := new(presenter.CollectionUser)
	resp, err := s.client.Do(ctx, req, &userCollResponse)
	if err != nil {
		return nil, resp, err
	}

	return userCollResponse, resp, nil
}

func (s *UserCollService) UpdateUserColl(ctx context.Context, collectionID int64, title string) (*presenter.CollectionUser, *http.Response, error) {
	payload := &user.NewUserCollection{
		Title: title,
	}

	req, err := s.client.NewRequest("PUT", fmt.Sprintf("/users/collections/%d/", collectionID), payload)
	if err != nil {
		return nil, nil, err
	}

	userCollResponse := new(presenter.CollectionUser)
	resp, err := s.client.Do(ctx, req, &userCollResponse)
	if err != nil {
		return nil, resp, err
	}

	return userCollResponse, resp, nil
}

func (s *UserCollService) DeleteUserColl(ctx context.Context, collectionID int64) (*http.Response, error) {
	req, err := s.client.NewRequest("DELETE", fmt.Sprintf("/users/collections/%d/", collectionID), nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

func (s *UserCollService) ListUserCollProd(ctx context.Context, collectionID int64) ([]*presenter.Product, *http.Response, error) {
	req, err := s.client.NewRequest("GET", fmt.Sprintf("/users/collections/%d/products/", collectionID), nil)
	if err != nil {
		return nil, nil, err
	}

	var products []*presenter.Product
	resp, err := s.client.Do(ctx, req, &products)
	if err != nil {
		return nil, resp, err
	}

	return products, resp, nil
}

func (s *UserCollService) CreateProdInUserColl(ctx context.Context, collectionID, productID int64) (*presenter.Product, *http.Response, error) {
	payload := user.NewUserCollectionProduct{
		ProductID: &productID,
	}

	req, err := s.client.NewRequest("POST", fmt.Sprintf("/users/collections/%d/products/", collectionID), payload)
	if err != nil {
		return nil, nil, err
	}

	product := new(presenter.Product)
	resp, err := s.client.Do(ctx, req, product)
	if err != nil {
		return nil, resp, err
	}

	return product, resp, nil
}

func (s *UserCollService) DeleteProdFromUserColl(ctx context.Context, collectionID, productID int64) (*http.Response, error) {
	req, err := s.client.NewRequest("DELETE", fmt.Sprintf("/products/%d/collections/%d/", productID, collectionID), nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

func (s *UserCollService) DeleteProdFromAllUserColl(ctx context.Context, productID int64) (*http.Response, error) {
	req, err := s.client.NewRequest("DELETE", fmt.Sprintf("/products/%d/collections/", productID), nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}