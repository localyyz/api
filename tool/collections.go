package tool

import (
	"log"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/htmlx"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	db "upper.io/db.v3"
)

func SyncDeals(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	client := ctx.Value("shopify.client").(*shopify.Client)

	collections, _, _ := client.CollectionList.List(ctx, nil)
	for _, c := range collections {
		dbCollection, err := data.DB.Collection.FindOne(db.Cond{
			"external_id": c.ID,
		})
		if err != nil {
			if err == db.ErrNoMoreRows {
				dbCollection = &data.Collection{
					Name:        c.Title,
					ExternalID:  &(c.ID),
					Lightning:   true,
					Description: htmlx.StripTags(c.BodyHTML),
					Status:      data.CollectionStatusQueued,
				}
				if c.Image != nil {
					dbCollection.ImageURL = c.Image.Src
				}
				data.DB.Collection.Save(dbCollection)

			}

			continue
		}

		// fetch the products
		productIDs, _, _ := client.CollectionList.ListProductIDs(ctx, c.ID)
		if len(productIDs) == 0 {
			continue
		}

		products, err := data.DB.Product.FindAll(db.Cond{"external_id": productIDs})
		for _, p := range products {
			err := data.DB.CollectionProduct.Create(data.CollectionProduct{
				CollectionID: dbCollection.ID,
				ProductID:    p.ID,
			})
			if err != nil {
				log.Println(err)
			}
		}
	}
}
