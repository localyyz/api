package reporter

import (
	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/stash"
)

func HandleProductViewed(product *data.Product) {
	stash.IncrProductViews(product.ID)
}

func HandleProductPurchased(product *data.Product) {
	stash.IncrProductPurchases(product.ID)
}
