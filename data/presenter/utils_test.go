package presenter

import (
	"testing"
)

type utilTest struct {
	name     string
	input    string
	expected interface{}
}

func TestGenerateThumbUrl(t *testing.T) {
	t.Parallel()

	tests := []utilTest{
		{
			name:     "general input",
			input:    "https://cdn.shopify.com/s/files/1/1734/8545/products/PS02_FRONT.jpg?v=1516572613",
			expected: "https://cdn.shopify.com/s/files/1/1734/8545/products/PS02_FRONT_medium.jpg?v=1516572613",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if actual := thumbImage(tt.input); actual != tt.expected {
				t.Errorf("test '%s': expected '%v', got '%v'", tt.name, tt.expected, actual)
			}
		})
	}
}
