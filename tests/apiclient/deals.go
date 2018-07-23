package apiclient

import (
	"context"
	"fmt"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/deals"
)

type DealService service

func (s *DealService) Activate(ctx context.Context, payload *deals.ActivateRequest) (*presenter.LightningCollection, *http.Response, error) {
	req, err := s.client.NewRequest("POST", "/deals/activate", payload)
	if err != nil {
		return nil, nil, err
	}

	dealResponse := new(presenter.LightningCollection)
	resp, err := s.client.Do(ctx, req, dealResponse)
	if err != nil {
		return nil, resp, err
	}

	return dealResponse, resp, nil
}

func (s *DealService) ListActive(ctx context.Context) ([]*presenter.LightningCollection, *http.Response, error) {
	req, err := s.client.NewRequest("GET", "/deals/active", nil)
	if err != nil {
		return nil, nil, err
	}

	var dealResponses []*presenter.LightningCollection
	resp, err := s.client.Do(ctx, req, &dealResponses)
	if err != nil {
		return nil, resp, err
	}

	return dealResponses, resp, nil
}

func (s *DealService) Get(ctx context.Context, dealID int64) (*presenter.LightningCollection, *http.Response, error) {
	req, err := s.client.NewRequest("GET", fmt.Sprintf("/deals/active/%d", dealID), nil)
	if err != nil {
		return nil, nil, err
	}

	dealResponse := new(presenter.LightningCollection)
	resp, err := s.client.Do(ctx, req, dealResponse)
	if err != nil {
		return nil, resp, err
	}

	return dealResponse, resp, nil
}
