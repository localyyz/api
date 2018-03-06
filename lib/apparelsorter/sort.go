package apparelsorter

import "sort"

type Size struct {
	Abbr  string
	Size  string
	Order int
}
type SizeSorter []*Size

func New(sizes ...string) SizeSorter {
	var l []*Size
	for _, s := range sizes {
		// match one of the size regex in
		if ss := matchSize(s); ss != nil {
			l = append(l, matchSize(s))
		}
	}
	return SizeSorter(l)
}

func Sort(sorter SizeSorter) {
	sort.Sort(sorter)
}

func (sort SizeSorter) Len() int {
	return len(sort)
}

func (sort SizeSorter) Swap(i, j int) {
	sort[i], sort[j] = sort[j], sort[i]
}

func (sort SizeSorter) Less(i, j int) bool {
	return sort[i].Order < sort[j].Order
}

func (sort SizeSorter) StringSlice() []string {
	slice := make([]string, len(sort))
	for i, s := range sort {
		slice[i] = s.Size
	}
	return slice
}
