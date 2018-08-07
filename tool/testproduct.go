package tool

import (
	"fmt"
	"log"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
)

// this is used to insert test product and test merchant that can be used to
// test end to end purchase.

const TestShopifyID = `best-test-store-toronto`
const TestAccessToken = `a5de53c040ee7c3f937d363d52858aeb`
const TestVariantID = 43252300547

func InsertTestPurchasableProduct(w http.ResponseWriter, r *http.Request) {
	// find/create the merchant by shopify_id
	place, err := data.DB.Place.FindOne(db.Cond{"shopify_id": TestShopifyID})
	if err != nil {
		if err == db.ErrNoMoreRows {
			place = &data.Place{
				ShopifyID: TestShopifyID,
			}
		} else {
			log.Fatal(err)
		}
	}
	place.Status = data.PlaceStatusActive
	place.Name = "LocalyyzTest Merchant"
	if err := data.DB.Place.Save(place); err != nil {
		log.Fatal(err)
	}
	lg.Warnf("inserted pl(id=%d) name: %s", place.ID, place.Name)

	// insert into shopify_credentials
	creds, err := data.DB.ShopifyCred.FindByPlaceID(place.ID)
	if err != nil {
		if err == db.ErrNoMoreRows {
			creds = &data.ShopifyCred{
				PlaceID:     place.ID,
				ApiURL:      fmt.Sprintf("https://%s.myshopify.com", TestShopifyID),
				AccessToken: TestAccessToken,
			}
		} else {
			log.Fatal(err)
		}
	}
	creds.Status = data.ShopifyCredStatusActive
	if err := data.DB.ShopifyCred.Save(creds); err != nil {
		log.Fatal(err)
	}

	// delete all products
	data.DB.Product.Find(db.Cond{"place_id": place.ID}).Delete()

	// insert testable product
	product := &data.Product{
		Title:   "localyyztest - sample product",
		Status:  data.ProductStatusApproved,
		PlaceID: place.ID,
		Score:   6,
	}
	if err := data.DB.Product.Save(product); err != nil {
		log.Fatal(err)
	}
	lg.Warnf("inserted product(id=%d) title: %s", product.ID, product.Title)
	pv := &data.ProductVariant{
		ProductID: product.ID,
		PlaceID:   place.ID,
		Price:     213.60,
		Limits:    999,
		OfferID:   TestVariantID,
		Etc: data.ProductVariantEtc{
			Size:  "small",
			Color: "deep",
		},
	}
	if err := data.DB.ProductVariant.Save(pv); err != nil {
		log.Fatal(err)
	}

	// product images
	img := &data.ProductImage{
		ProductID:  product.ID,
		ImageURL:   "https://cdn.shopify.com/s/files/1/1976/6885/products/01292015_Ashley_01_21092_1345.jpg?v=1502237757",
		Score:      1,
		ExternalID: 23964498307,
	}
	if err := data.DB.ProductImage.Save(img); err != nil {
		log.Fatal(err)
	}

}
