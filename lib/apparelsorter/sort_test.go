package apparelsorter

import (
	"testing"
)

type sizeTest struct {
	name     string
	inputs   []string
	expected []*Size
}

func TestSizeSorter(t *testing.T) {
	t.Parallel()

	tests := []sizeTest{
		{"l m s", []string{"l", "s", "m"}, []*Size{{Size: "s"}, {Size: "m"}, {Size: "l"}}},
		{"l m xs", []string{"l", "xs", "m"}, []*Size{{Size: "xs"}, {Size: "m"}, {Size: "l"}}},
		{"large medium small", []string{"large", "small", "medium"}, []*Size{{Size: "small"}, {Size: "medium"}, {Size: "large"}}},
		{"extra small", []string{"large", "small", "extra small"}, []*Size{{Size: "extra small"}, {Size: "small"}, {Size: "large"}}},
		{"extra large", []string{"large", "small", "extra large"}, []*Size{{Size: "small"}, {Size: "large"}, {Size: "extra large"}}},
		{"xxlarge", []string{"medium", "small", "xxlarge"}, []*Size{{Size: "small"}, {Size: "medium"}, {Size: "xxlarge"}}},
		{"xxl", []string{"m", "s", "xxl"}, []*Size{{Size: "s"}, {Size: "m"}, {Size: "xxl"}}},
		{"2xl", []string{"2xl", "l", "xl"}, []*Size{{Size: "l"}, {Size: "xl"}, {Size: "2xl"}}},
		{"5 8 10", []string{"5", "10", "8"}, []*Size{{Size: "5"}, {Size: "8"}, {Size: "10"}}},
		{"5 8 10.0", []string{"5", "10.0", "8"}, []*Size{{Size: "5"}, {Size: "8"}, {Size: "10.0"}}},
		{"5.5 6 6.5", []string{"6.5", "5.5", "6", "7", "5"}, []*Size{{Size: "5"}, {Size: "5.5"}, {Size: "6"}, {Size: "6.5"}, {Size: "7"}}},
		{"shoe size with string", []string{"5-men", "10-men", "8-men"}, []*Size{{Size: "5-men"}, {Size: "8-men"}, {Size: "10-men"}}},
		{"shoe size with string 2", []string{"women-5", "women-10", "women-8"}, []*Size{{Size: "women-5"}, {Size: "women-8"}, {Size: "women-10"}}},
		{"shoe size with string mixed float", []string{"women-5.5", "women-5", "women-8"}, []*Size{{Size: "women-5"}, {Size: "women-5.5"}, {Size: "women-8"}}},
		{"mixed eu and us", []string{"40 eur - 7 us", "39 eur - 6 us", "41 eur - 8 us"}, []*Size{{Size: "39 eur - 6 us"}, {Size: "40 eur - 7 us"}, {Size: "41 eur - 8 us"}}},
		{"unmatched size", []string{"one-size"}, []*Size{{Size: "one-size", Order: postpendIdx}}},
		{"mixed with unmatched size", []string{"some size", "m", "l"}, []*Size{{Size: "m"}, {Size: "l"}, {Size: "some size", Order: postpendIdx}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := New(tt.inputs...)
			Sort(actual)
			tt.compare(t, actual)
		})
	}
}

func (tt sizeTest) compare(t *testing.T, actual []*Size) {
	if len(tt.expected) != len(actual) {
		t.Fatalf("test '%s': expected list length '%d', got '%d'", tt.name, len(tt.expected), len(actual))
	}

	for i, e := range tt.expected {
		if e.Size != actual[i].Size {
			t.Errorf("test '%s': expected size '%s' at %d, got '%s'", tt.name, e.Size, i, actual[i].Size)
		}
		if e.Order != 0 && e.Order != actual[i].Order {
			t.Errorf("test '%s': expected order '%d', got '%d'", tt.name, e.Order, actual[i].Order)
		}
	}
}
