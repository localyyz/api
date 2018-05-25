package sync

import (
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
)

type productImageSyncer interface {
	FetchProductImages() ([]*data.ProductImage, error)
	GetProduct() *data.Product
	Finalize([]*data.ProductImage, []*data.ProductImage) error
	ValidateImages() bool
}

type productScorer interface {
	GetProductImages(int64) ([]*data.ProductImage, error)
	GetProduct() *data.Product
	GetPlace() *data.Place
	CheckPriority() bool
	Finalize([]*data.ProductImage) error
}

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}
