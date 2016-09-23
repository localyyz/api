package presenter

import (
	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/goware/geotools"
)

type User struct {
	*data.User

	Geo    geotools.Point `json:"geo"`
	Locale *data.Locale   `json:"locale"`
}
