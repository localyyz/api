package sync

import (
	"testing"
)

func TestSetPrices(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		price         string
		comparePrice  string
		expectedPrice float64
		expectedPrev  float64
	}{
		{
			name:          "price only",
			price:         "10.00",
			expectedPrice: 10.00,
		},
		{
			name:          "compare at price > price (discount)",
			price:         "50.00",
			comparePrice:  "100.00",
			expectedPrev:  100.00,
			expectedPrice: 50.00,
		},
		{
			name:          "compare at price = price (discount)",
			price:         "50.00",
			comparePrice:  "50.00",
			expectedPrice: 50.00,
		},
		{
			name:          "compare at price < price (invalid)",
			price:         "50.00",
			comparePrice:  "10.00",
			expectedPrice: 0.00,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			price, prevPrice := setPrices(tt.price, tt.comparePrice)
			if price != tt.expectedPrice {
				t.Errorf("test '%s': expected '%.2f', got '%.2f'", tt.name, tt.expectedPrice, price)
			}
			if prevPrice != tt.expectedPrev {
				t.Errorf("test '%s': expected '%.2f', got '%.2f'", tt.name, tt.expectedPrev, prevPrice)
			}
		})
	}
}
