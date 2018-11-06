package sync

import (
	"reflect"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	xchange "bitbucket.org/moodie-app/moodie-api/lib/xchanger"
	"github.com/pkg/errors"
	set "gopkg.in/fatih/set.v0"
)

var (
	ErrProductUnavailable  = errors.New("unavailable")
	ErrProductInvalidPrice = errors.New("invalid price")
)

type shopifyVariantSyncer struct {
	product   *data.Product
	place     *data.Place
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
	if len(s.toSaves) == 0 {
		// this means for whatever the reason, all variants
		// were removed
		return ErrProductUnavailable
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
	managedInventory := true
	for _, v := range variants {
		if v.InventoryManagement != "shopify" {
			managedInventory = false
			break
		}
		inventorySum += v.InventoryQuantity
	}
	// some merchants DO NOT automatically manage inventory
	// via shopify. Only mark product unavailable if inventory
	// is managed by shopify
	if inventorySum <= 0 && managedInventory {
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
		if price == 0 {
			return ErrProductInvalidPrice
		}

		if i == 0 {
			s.product.Price = xchange.ToUSD(price, s.place.Currency)
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
			n := strings.ToLower(o.Name)
			vv := strings.ToLower(o.Value)
			// comment: the naming for sizes are fairly complex.
			// for example, 'it/men 46'
			if strings.Contains(n, "size") {
				etc.SizeName = n
				etc.Size = vv
			} else if strings.Contains(n, "color") {
				etc.Color = vv
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

		// some merchants DO NOT automatically manage inventory
		// via shopify. Only mark product unavailable if inventory
		// is managed by shopify
		// NOTE: hard code inventory to 999
		if !managedInventory {
			//
			editV.Limits = 999
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
