package sync

import (
	"reflect"
	"sort"
	"testing"
)

func TestParseTags(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		tagStr   string
		optTags  []string
		expected []string
	}{
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := parseTags(tt.tagStr, tt.optTags...)
			sort.Strings(actual)
			if !reflect.DeepEqual(tt.expected, actual) {
				t.Errorf("test %s: got %+v", tt.name, actual)
			}
		})
	}
}
