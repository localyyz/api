package reporter

import (
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/data/stash"
)

func HandleProductViewed(product *presenter.ProductEvent) {
	stash.IncrProductViews(product.ID, product.ViewerID)
}

func HandleProductPurchased(product *presenter.ProductEvent) {
	stash.IncrProductPurchases(product.ID)
}
