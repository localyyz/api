package shopify

import (
	"context"
	"net/http"
)

type ProductListService service

type ProductList Product

func (p *ProductListService) Get(ctx context.Context) ([]*ProductList, *http.Response, error) {
	req, err := p.client.NewRequest("GET", "/admin/product_listings.json", nil)
	if err != nil {
		return nil, nil, err
	}

	var productListWrapper struct {
		ProductListings []*ProductList `json:"product_listings"`
	}
	resp, err := p.client.Do(ctx, req, &productListWrapper)
	if err != nil {
		return nil, resp, err
	}

	return productListWrapper.ProductListings, resp, nil
}
