package shopify

import (
	"encoding/json"
	"strings"
	"testing"
)

type errorTest struct {
	name     string
	input    string
	expected string
}

func TestShopifyError(t *testing.T) {
	t.Parallel()

	inputs := []errorTest{
		{
			name: "line item invalid variant_id",
			input: `
{
  "errors": {
    "line_items": {
      "0": {
        "variant_id": [
          {
            "code": "invalid",
            "message": "is invalid",
            "options": {}
          }
        ]
      }
    }
  }
}`,
			expected: "line_items at pos(0): variant_id is invalid",
		},
		{
			name: "invalid blank shipping zip",
			input: `
{
  "errors": {
    "shipping_address": {
      "zip": [
        {
          "code": "blank",
          "message": "can't be blank",
          "options": {}
        }
      ]
    }
  }
}`,
			expected: "shipping_address: zip can't be blank",
		},
		{
			name: "invalid blank billing zip",
			input: `
{
  "errors": {
    "billing_address": {
      "zip": [
        {
          "code": "blank",
          "message": "can't be blank",
          "options": {}
        }
      ]
    }
  }
}`,
			expected: "billing_address: zip can't be blank",
		},
	}

	for _, tt := range inputs {
		r := &ErrorResponse{}
		json.Unmarshal([]byte(tt.input), &r)
		actual := findFirstError(r)
		if !strings.Contains(actual.Error(), tt.expected) {
			t.Errorf("%s: expected %s got %s", tt.name, tt.expected, actual)
		}
	}

}
