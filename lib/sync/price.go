package sync

import (
	"math"
	"strconv"
)

// setPrices helper compares product.price and product.comparePrice
// and always sets the lower one to the `price` field and higher one to `prevPrice`
// field.
func setPrices(priceStr, compareStr string) (price, comparePrice float64) {
	price1, _ := strconv.ParseFloat(priceStr, 64)
	if len(compareStr) == 0 {
		price = price1
		return
	}
	price2, _ := strconv.ParseFloat(compareStr, 64)
	if price2 > 0 && price1 > price2 {
		//price = price2
		//comparePrice = price1

		// some how, compare at price is invalid
		return price1, 0
	} else if price2 > 0 && price2 > price1 {
		price = price1
		comparePrice = price2
	} else {
		price = price1
	}
	return
}

func pctRound(value float64, precision float64) float64 {
	multiplier := math.Pow(10.0, precision)
	return math.Round((1.0-value)*multiplier) / multiplier
}
