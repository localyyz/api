package presenter

import "bitbucket.org/moodie-app/moodie-api/data"

type Promo struct {
	*data.Promo

	//fields that can be viewed
	NumClaimed int64 `json:"numClaimed"`
}
