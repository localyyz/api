package presenter

import (
	"context"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/go-chi/render"
)

type ShippingZone struct {
	*data.ShippingZone

	//Region     string `json:"region"`
	//RegionCode string `json:"regionCode"`

	//Regions    interface{} `json:"regions,omitempty"`
	PlaceID    interface{} `json:"placeID,omitempty"`
	ExternalID interface{} `json:"externalID,omitempty"`

	ctx context.Context
}

func NewShippingZoneList(ctx context.Context, zones []*data.ShippingZone) []render.Renderer {
	list := []render.Renderer{}
	for _, zone := range zones {
		list = append(list, &ShippingZone{ShippingZone: zone})
	}
	return list
}

// Place implements render.Renderer interface
func (*ShippingZone) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
