package tool

import (
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
		// check if collection already exists
		exists, _ := data.DB.Collection.Find(db.Cond{
			"external_id": c.ID,
		}).Exists()
		if exists {
			continue
		}

		dbCollection := &data.Collection{
			Name:        c.Title,
			ExternalID:  &(c.ID),
			Lightning:   true,
			Description: htmlx.StripTags(c.BodyHTML),
		}
		if c.Image != nil {
			dbCollection.ImageURL = c.Image.Src
		}
		data.DB.Collection.Save(dbCollection)

		// fetch the products
		productIDs, _, _ := client.CollectionList.ListProductIDs(ctx, c.ID)

		products, _ := data.DB.Product.FindAll(db.Cond{"external_id": productIDs})
		for _, p := range products {
			data.DB.CollectionProduct.Save(&data.CollectionProduct{
				CollectionID: dbCollection.ID,
				ProductID:    p.ID,
			})
		}
	}
}
