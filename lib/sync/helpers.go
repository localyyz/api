package sync

import (
	"strconv"
)

// setPrices helper compares product.price and product.comparePrice
// and always sets the lower one to the `price` field and higher one to `prevPrice`
// field.
func setPrices(a, b string) (price, comparePrice float64) {
	price1, _ := strconv.ParseFloat(a, 64)
	if len(b) == 0 {
		price = price1
		return
	}
	price2, _ := strconv.ParseFloat(b, 64)
	if price2 > 0 && price1 > price2 {
		price = price2
		comparePrice = price1
	} else if price2 > 0 && price2 > price1 {
		price = price1
		comparePrice = price2
	} else {
		price = price1
	}
	return
}