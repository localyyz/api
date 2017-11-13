package shopify

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

type ProductListService service

type ProductList Product

type ProductListParam struct {
	ProductIDs   []int64
	CollectionID int64
	Handle       string
	Limit        int
	Page         int
	UpdatedAtMin time.Time
}

func (p *ProductListParam) EncodeQuery() string {
	if p == nil {
		return ""
	}
	// for now just allow handle
	// TODO: support all params
	v := url.Values{}
	v.Add("handle", p.Handle)
	return v.Encode()
}

func (p *ProductListService) Get(ctx context.Context, params *ProductListParam) ([]*ProductList, *http.Response, error) {
	req, err := p.client.NewRequest("GET", "/admin/product_listings.json", nil)
	if err != nil {
		return nil, nil, err
	}
	// encode param to query
	req.URL.RawQuery = params.EncodeQuery()

	var productListWrapper struct {
		ProductListings []*ProductList `json:"product_listings"`
	}
	resp, err := p.client.Do(ctx, req, &productListWrapper)
	if err != nil {
		return nil, resp, err
	}

	return productListWrapper.ProductListings, resp, nil
}
