package sync

import (
	"reflect"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"github.com/pkg/errors"
	set "gopkg.in/fatih/set.v0"
)

var (
	ErrProductUnavailable = errors.New("unavailable")
)

type shopifyVariantSyncer struct {
	product   *data.Product
	toSaves   []*data.ProductVariant
	toRemoves []*data.ProductVariant
}

func (s *shopifyVariantSyncer) Finalize() error {
	for _, v := range s.toSaves {
		data.DB.ProductVariant.Save(v)
	}
	for _, v := range s.toRemoves {
		data.DB.ProductVariant.Delete(v)
	}
	return nil
}

// fetches existing product images from the database.
// function is abstracted so the db call can be mocked/tested
func (s *shopifyVariantSyncer) FetchFromDB() ([]*data.ProductVariant, error) {
	return data.DB.ProductVariant.FindByProductID(s.product.ID)
}

func (s *shopifyVariantSyncer) Sync(variants []*shopify.ProductVariant) error {
	inventorySum := 0
	for _, v := range variants {
		inventorySum += v.InventoryQuantity
	}
	if inventorySum == 0 {
		return ErrProductUnavailable
	}

	//getting the images from product_images for product
	dbVariants, err := s.FetchFromDB()
	if err != nil {
		return err
	}

	//fill out a map using external image IDS
	dbVariantsMap := map[int64]*data.ProductVariant{}
	for _, v := range dbVariants {
		dbVariantsMap[v.OfferID] = v
	}

	syncVariantsSet := set.New()

	for i, v := range variants {
		// add to variants sets.
		syncVariantsSet.Add(v.ID)

		price, prevPrice := setPrices(
			v.Price,
			v.CompareAtPrice,
		)

		if i == 0 {
			s.product.Price = price
			if price > 0 && prevPrice > price {
				s.product.DiscountPct = pctRound(price/prevPrice, 1)
				// if discounted_at is not set. set it.
				if s.product.DiscountedAt == nil {
					s.product.DiscountedAt = data.GetTimeUTCPointer()
				}
			} else {
				s.product.DiscountPct = 0
				s.product.DiscountedAt = nil
			}
		}

		etc := data.ProductVariantEtc{Sku: v.Sku}
		// variant option values
		for _, o := range v.OptionValues {
			vv := strings.ToLower(o.Value)
			switch strings.ToLower(o.Name) {
			case "size":
				etc.Size = vv
			case "color":
				etc.Color = vv
			default:
				// pass
			}
		}

		editV := &data.ProductVariant{
			ProductID:   s.product.ID,
			PlaceID:     s.product.PlaceID,
			Limits:      int64(v.InventoryQuantity),
			Description: v.Title,
			OfferID:     v.ID,
			Price:       price,
			PrevPrice:   prevPrice,
			Etc:         etc,
		}

		// check if variant has changed from the one in db
		if dbV, ok := dbVariantsMap[v.ID]; ok {
			editV.ID = dbV.ID
			if reflect.DeepEqual(editV, dbV) {
				continue
			}
		}

		s.toSaves = append(s.toSaves, editV)
	}

	for _, v := range dbVariantsMap {
		if !syncVariantsSet.Has(v.OfferID) {
			s.toRemoves = append(s.toRemoves, v)
		}
	}

	return s.Finalize()
}
