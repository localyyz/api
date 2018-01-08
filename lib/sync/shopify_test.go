package sync

import (
	"encoding/json"
	"reflect"
	"sort"
	"testing"
)

type tagTest struct {
	name     string
	tagStr   string
	optTags  []string
	expected []string
}

func TestParseTags(t *testing.T) {
	t.Parallel()
	tests := []tagTest{
		{
			name:     "simple plural category",
			tagStr:   "dresses",
			expected: []string{"dress"},
		},
		{
			name:     "simple plural category with gender",
			tagStr:   "women's dresses",
			expected: []string{"dress", "woman"},
		},
		{
			name:     "complex category with gender",
			tagStr:   "women's coats & jackets",
			expected: []string{"coat", "jacket", "woman"},
		},
		{
			name:     "dash separators allowed",
			tagStr:   "men's button-ups",
			expected: []string{"button-up", "man"},
		},
		{
			name:     "dash separators allowed v2",
			tagStr:   "men's t-shirts",
			expected: []string{"man", "t-shirt"},
		},
		{
			name:     "dash space dash",
			tagStr:   "button up - long sleeve",
			expected: []string{"button up", "long sleeve"},
		},
		{
			name:     "complicated dashes",
			tagStr:   "gift set - 3.3 oz eau de toilette spray + 3.3 oz body lotion",
			expected: []string{"3.3 oz eau de toilette spray", "3.3 oz body lotion", "gift set"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := parseTags(tt.tagStr, tt.optTags...)
			tt.compareStringSlice(t, actual)
		})
	}
}

func (tt tagTest) compareStringSlice(t *testing.T, actual []string) {
	sort.Strings(actual)
	if !reflect.DeepEqual(tt.expected, actual) {
		out, _ := json.Marshal(actual)
		in, _ := json.Marshal(tt.expected)
		t.Errorf("test %s: expected %s got %s", tt.name, string(in), string(out))
	}
}
