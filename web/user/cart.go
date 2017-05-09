package user

import "net/http"

// TODO: shopping list concept
// claim -> product -> place is too complicated
//
// The architecture should be:
// - products can be added to shopping carts
//    at "checkout" pick the promotion if available
// - multiple shopping carts? would that become "collections"?
func GetCart(w http.ResponseWriter, r *http.Request) {

}
