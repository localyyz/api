package sync

import (
	"testing"

	"bitbucket.org/moodie-app/moodie-api/data"
)

func TestSetPrices(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		price        string
		comparePrice string
		expected     data.ProductVariantEtc
	}{
		{
			name:     "price only",
			price:    "10.00",
			expected: data.ProductVariantEtc{Price: 10.00},
		},
		{
			name:         "compare at price > price (discount)",
			price:        "50.00",
			comparePrice: "100.00",
			expected:     data.ProductVariantEtc{PrevPrice: 100.00, Price: 50.00},
		},
		{
			name:         "compare at price = price (discount)",
			price:        "50.00",
			comparePrice: "50.00",
			expected:     data.ProductVariantEtc{Price: 50.00},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := setPrices(tt.price, tt.comparePrice)
			if actual != tt.expected {
				t.Errorf("test '%s': expected '%+v', got '%+v'", tt.name, tt.expected, actual)
			}
		})
	}
}
