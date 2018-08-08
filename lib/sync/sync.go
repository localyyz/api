package sync

import (
	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"gopkg.in/fatih/set.v0"
	"strings"
	db "upper.io/db.v3"
)

const (
	LOCALYYDOTDSTRING = "LOCALYYDOTD"
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

func (s *productSyncer) SyncVariants(variants []*shopify.ProductVariant) error {
	if err := (&shopifyVariantSyncer{product: s.product}).Sync(variants); err != nil {
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

///////
// SOMEWHAT HACKY
// WE NEED TO MARK CERTAIN PRODUCTS AS DOTD FOR BUYING AFTER/BEFORE DOTD ENDS -> LOOK FOR THE TAG FROM THE SHOPIFY STORE
func (s *productSyncer) SyncDOTDProductStatus(tags string) {
	splitTags := strings.Split(tags, " ")
	tagSet := set.New()
	for _, tag := range splitTags {
		tagSet.Add(tag)
	}
	if tagSet.Has(LOCALYYDOTDSTRING) {
		s.product.Status = data.ProductStatusDOTD
	}
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
