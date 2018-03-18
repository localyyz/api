package data

import (
	"github.com/pressly/lg"
	"upper.io/bond"
	db "upper.io/db.v3"
)

type SearchWord struct {
	Word string `db:"word" json:"word"`
	// # of document occurrences
	NDoc int64 `db:"ndoc" json:"nDoc"`

	// Filler field for queries
	Similarity float64 `db:"similarity" json:"similarity"`
}

type SearchWordStore struct {
	bond.Store
}

func (*SearchWord) CollectionName() string {
	return `search_words`
}

func (store SearchWordStore) FindSimilar(words ...string) ([]*SearchWord, error) {
	var searchWords []*SearchWord

	for _, s := range words {
		var sw *SearchWord
		// NOTE magic number 0.3 is a similarity threshold, by default it's what
		// the operator %> returns. however there's some weirdness where results
		// are incorrectly filtered here. So we use `similarity` instead, which
		// is the samething without the default 0.3 filter

		// we also filter for length of words between +/-1 of the search word
		query := store.
			Find(
				db.Raw("similarity(word, ?) > 0.3", s),
				db.Raw("length(word) BETWEEN ? AND ?", len(s)-1, len(s)+1),
			).
			Select("*", db.Raw("similarity(word, ?) AS similarity", s)).
			OrderBy(db.Raw("similarity(word, ?) DESC", s), "ndoc")
		if err := query.One(&sw); err != nil {
			lg.Warnf("invalid search word '%s' with %v", s, err)
			continue
		}
		searchWords = append(searchWords, sw)
	}
	return searchWords, nil
}

type WordFrequencySorter []*SearchWord

func (sort WordFrequencySorter) Len() int {
	return len(sort)
}

func (sort WordFrequencySorter) Swap(i, j int) {
	sort[i], sort[j] = sort[j], sort[i]
}

func (sort WordFrequencySorter) Less(i, j int) bool {
	return sort[i].Similarity > sort[j].Similarity || sort[i].NDoc < sort[j].NDoc
}
