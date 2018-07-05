package reporter

import (
	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/data/stash"
)

func HandleProductViewed(product *presenter.ProductView) {
	stash.IncrProductViews(product.ID, product.ViewerID)
}

func HandleProductPurchased(product *data.Product) {
	stash.IncrProductPurchases(product.ID)
}
