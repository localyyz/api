package sync

import "bitbucket.org/moodie-app/moodie-api/data"

type productImageSyncer interface {
	FetchProductImages() ([]*data.ProductImage, error)
	GetProduct() *data.Product
	Finalize([]*data.ProductImage, []*data.ProductImage) error
}