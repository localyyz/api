package sync

import (
	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	db "upper.io/db.v3"
)

type productSyncer struct {
	product  *data.Product
	place    *data.Place
	variants []*data.ProductVariant
	images   []*data.ProductImage
}

type Syncer interface {
	Sync() error
}

type Finalizer interface {
	Finalize() error
}

type Fetcher interface {
	Fetch(db.Cond, interface{}) error
}

// retry syncing
func (s *productSyncer) Retry() {
}

func (s *productSyncer) SyncCategories(title, tags, productType string) error {
	catSync := &shopifyCategorySyncer{
		product: s.product,
		place:   s.place,
	}
	if err := catSync.Sync(title, tags, productType); err != nil {
		if err == ErrBlacklisted {
			// rejected. product category is blacklisted
			s.FinalizeStatus(data.ProductStatusRejected)
		}
		// TODO: if error is detected. retry?
		return err
	}
	return nil
}

func (s *productSyncer) SyncVariants(variants []*shopify.ProductVariant) error {
	if err := (&shopifyVariantSyncer{product: s.product, place: s.place}).Sync(variants); err != nil {
		if err == ErrProductUnavailable {
			// rejected. no inventory quantity
			s.FinalizeStatus(data.ProductStatusOutofStock)
		}
		// TODO: if error is detected. retry?
		return err
	}
	return nil
}

func (s *productSyncer) SyncImages(images []*shopify.ProductImage) error {
	if err := (&shopifyImageSyncer{Product: s.product}).Sync(images); err != nil {
		if err == ErrInvalidImage {
			// rejected. one or more images are invalid
			s.FinalizeStatus(data.ProductStatusRejected)
		}
		// TODO: if error is detected. retry?
		return err
	}
	return nil
}

func (s *productSyncer) SyncScore() error {
	score, err := (&shopifyProductScorer{Product: s.product, Place: s.place}).GetScore()
	if err != nil {
		return err
	}
	s.product.Score = score
	return nil
}

func (s *productSyncer) FinalizeStatus(status data.ProductStatus) error {
	s.product.Status = status
	return s.Finalize()
}

func (s *productSyncer) Finalize() error {
	// product was not previously marked as pending.
	if s.product.Status == data.ProductStatusProcessing {
		s.product.Status = data.ProductStatusApproved
	}
	data.DB.Product.Save(s.product)
	return nil
}
