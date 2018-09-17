package events

import "sync"

type Event string

var evtMap map[string]Event
var evtLock sync.Mutex

const (
	EvProduct            = "data.product"
	EvProductViewed      = "data.product.viewed"
	EvProductPurchased   = "data.product.purchased"
	EvProductAddedToCart = "data.product.addedtocart"
	EvProductFavourited  = "data.product.favourited"
)

func init() {
	evtMap = make(map[string]Event)

	RegisterEvent(string(EvProductViewed), EvProductViewed)
	RegisterEvent(string(EvProductPurchased), EvProductPurchased)
	RegisterEvent(string(EvProductFavourited), EvProductFavourited)
	RegisterEvent(string(EvProductAddedToCart), EvProductAddedToCart)
}

// EventForType will return the registered Event for the evtType.
func EventForType(evtType string) Event {
	evtLock.Lock()
	defer evtLock.Unlock()
	return evtMap[evtType]
}

// RegisterEvent will register the evtType with the given Event. Useful for customization.
func RegisterEvent(evtType string, evt Event) {
	evtLock.Lock()
	defer evtLock.Unlock()
	evtMap[evtType] = evt
}
