package tool

import (
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/scheduler"
)

const LocalyyzStoreId = 4164
const DotdCollectionId = 76596346998

func SyncDeals(w http.ResponseWriter, r *http.Request) {

	h := scheduler.New(nil)

	h.SyncDOTD()

	h.Wait()
}
