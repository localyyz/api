package sync

import (
	"context"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
)

type productSyncer struct {
	product  *data.Product
	place    *data.Place
	variants []*data.ProductVariant
	images   []*data.ProductImage

	listener Listener
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

func NewSyncer(ctx context.Context, product *data.Product, place *data.Place) (*productSyncer, error) {
	// TODO: make opts
	listener, _ := ctx.Value(SyncListenerCtxKey).(Listener)
	syncer := &productSyncer{
		product:  product,
		place:    place,
		listener: listener,
	}
	return syncer, nil
}

func (s *productSyncer) Sync(sy *shopify.ProductList) error {
	go func() {
		if s.listener != nil {
			// inform caller that we're done
			defer func() { s.listener <- 1 }()
		}
		if err := s.SyncCategories(sy.Title, sy.Tags, sy.ProductType); err != nil {
			lg.Warnf("shopify sync categories: %v", err)
			return
		}
		if err := s.SyncVariants(sy.Variants); err != nil {
			lg.Warnf("shopify add variant: %v", err)
			return
		}
		if err := s.SyncImages(sy.Images); err != nil {
			lg.Warnf("shopify add images: %v", err)
			return
		}
		if err := s.SyncScore(); err != nil {
			lg.Warnf("shopify shopify score: %v", err)
			return
		}
		if err := s.Finalize(); err != nil {
			lg.Warnf("shopify finalize: %v", err)
			return
		}
	}()
	return nil
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
			if err := s.FinalizeStatus(data.ProductStatusRejected); err != nil {
				return err
			}
		}
		// TODO: if error is detected. retry?
		return err
	}
	return nil
}

func (s *productSyncer) SyncVariants(variants []*shopify.ProductVariant) error {
	if err := (&shopifyVariantSyncer{product: s.product}).Sync(variants); err != nil {
		if err == ErrProductUnavailable {
			// rejected. no inventory quantity
			if err := s.FinalizeStatus(data.ProductStatusOutofStock); err != nil {
				return err
			}
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
			if err := s.FinalizeStatus(data.ProductStatusRejected); err != nil {
				return err
			}
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
	return data.DB.Product.Save(s.product)
}
